const go = new Go();
WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((result) => {
    go.run(result.instance);
}).catch((err) => {
    console.error(err);
});

document.getElementById('convertForm').addEventListener('submit', async (event) => {
    event.preventDefault();

    const mapIdInput = document.getElementById('mapId');
    const audioFileInput = document.getElementById('audioFile');
    const mapFileInput = document.getElementById('mapFile');

    for (const input of [mapIdInput, audioFileInput, mapFileInput]) {
        input.classList.remove('is-invalid');
    }

    const hasMapId = mapIdInput.value.trim() !== '';
    const hasAudioFile = audioFileInput.files.length > 0;
    const hasMapFile = mapFileInput.files.length > 0;

    if ((!hasMapId && !hasMapFile) || (hasAudioFile && !hasMapFile)) {
        mapIdInput.classList.add('is-invalid');
        return;
    } else if (!hasMapId && !hasAudioFile) {
        audioFileInput.classList.add('is-invalid');
        return;
    }

    try {
        setLoading(true);

        const mapId = mapIdInput.value.trim();
        const useBeta = document.getElementById('useBeta').checked;
        const audioFile = audioFileInput.files[0];
        const mapFile = mapFileInput.files[0];

        const sbMap = await getSparebeatMap(mapId, mapFile, useBeta);
        const osuMap = convertSparebeatMap(JSON.stringify(sbMap));
        if (!osuMap.success) {
            throw new Error(osuMap.error || 'Unknown conversion error.');
        }
        const audioData = await getAudioData(mapId, audioFile, useBeta);
        const backgroundData = await createBackgroundImage(sbMap);
        const oszFile = await createOszFile(osuMap, audioData, backgroundData);

        const fileName = `${osuMap.metadata.artist} - ${osuMap.metadata.title}`;

        document.getElementById('downloadButton').onclick = () => {
            const a = document.createElement('a');
            a.href = URL.createObjectURL(oszFile);
            a.download = `${fileName}.osz`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
        };

        document.getElementById('resultContent').innerHTML = `Successfully converted <strong>${fileName}</strong>`;
        document.getElementById('result').classList.remove('d-none');
    } catch (err) {
        console.error(err);
        document.getElementById('errorInfo').textContent = err.message;
        document.getElementById('errorInfo').classList.remove('d-none');
    } finally {
        setLoading(false);
    }
});

function setLoading(isLoading) {
    if (isLoading) {
        document.getElementById('loading').classList.remove('d-none');
        document.getElementById('buttonText').textContent = 'Converting Map...';
        document.getElementById('convertButton').disabled = true;
        document.getElementById('errorInfo').classList.add('d-none');
        document.getElementById('errorInfo').textContent = '';
        document.getElementById('result').classList.add('d-none');
        document.getElementById('resultContent').innerHTML = '';
    } else {
        document.getElementById('loading').classList.add('d-none');
        document.getElementById('buttonText').textContent = 'Convert Map';
        document.getElementById('convertButton').disabled = false;
    }
}

async function getSparebeatMap(mapId, mapFile, useBeta) {
    if (mapFile) {
        const mapContent = await mapFile.text();
        try {
            return JSON.parse(mapContent);
        } catch (err) {
            throw new Error(`Invalid map file: ${err.message}`);
        }
    }

    const mapURL = useBeta
        ? `https://corsproxy.io/?url=https://beta.sparebeat.com/api/tracks/${mapId}/map` // bypass cors restrictions
        : `https://sparebeat.com/play/${mapId}/map`;

    try {
        const res = await fetch(mapURL);
        if (res.status === 404) throw new Error(`Map with ID "${mapId}" not found.`);
        if (!res.ok) throw new Error(`Failed to fetch map: ${res.status} ${res.statusText}`);
        return await res.json();
    } catch (err) {
        throw new Error(`Error fetching map: ${err.message}`);
    }
}

async function getAudioData(mapId, audioFile, useBeta) {
    if (audioFile) {
        return new Uint8Array(await audioFile.arrayBuffer());
    }

    const audioURL = useBeta
        ? `https://corsproxy.io/?url=https://beta.sparebeat.com/api/tracks/${mapId}/audio` // bypass cors restrictions
        : `https://sparebeat.com/play/${mapId}/music`;

    try {
        const res = await fetch(audioURL);
        if (res.status === 404) throw new Error(`Map with ID "${mapId}" not found.`);
        if (!res.ok) throw new Error(`Failed to fetch audio: ${res.status} ${res.statusText}`);
        return new Uint8Array(await res.arrayBuffer());
    } catch (err) {
        throw new Error(`Error fetching audio: ${err.message}`);
    }
}

async function createBackgroundImage(sbMap) {
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');
    canvas.width = 1920;
    canvas.height = 1080;

    const backgroundImg = new Image();
    backgroundImg.src = './images/background.png';
    await new Promise((resolve, reject) => {
        backgroundImg.onload = resolve;
        backgroundImg.onerror = (err) => reject(new Error(`Error loading background image: ${err.message || err}`));
    });
    ctx.drawImage(backgroundImg, 0, 0, canvas.width, canvas.height);

    const gradient = ctx.createLinearGradient(0, 0, 0, canvas.height);
    gradient.addColorStop(0, sbMap.bgColor?.[0] ?? '#43c6ac');
    gradient.addColorStop(1, sbMap.bgColor?.[1] ?? '#191654');
    ctx.globalAlpha = 0.8;
    ctx.fillStyle = gradient;
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    return new Promise((resolve) => {
        canvas.toBlob((blob) => {
            const reader = new FileReader();
            reader.onload = (event) => resolve(new Uint8Array(event.target.result));
            reader.readAsArrayBuffer(blob);
        }, 'image/png');
    });
}

async function createOszFile(osuMap, audioData, backgroundData) {
    const files = Object.entries(osuMap.files).map(([fileName, content]) => ({
        name: fileName,
        content: new TextEncoder().encode(content)
    }));

    files.push({ name: "audio.mp3", content: audioData });
    files.push({ name: "background.png", content: backgroundData });

    const zip = new JSZip();
    files.forEach(file => zip.file(file.name, file.content));

    return zip.generateAsync({
        type: "blob",
        compression: "DEFLATE",
        compressionOptions: { level: 6 }
    });
}