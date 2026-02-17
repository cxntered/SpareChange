const CORS_PROXY = "https://sparechange.cxntered.workers.dev/?";
const SPAREBEAT_URL_REGEX = /(beta\.)?sparebeat\.com\/play\/([^/]+)/;

const form = document.getElementById('convertForm');
const mapUrl = document.getElementById('mapUrl');
const mapUrlFeedback = document.getElementById('mapUrlFeedback');
const audioFile = document.getElementById('audioFile');
const mapFile = document.getElementById('mapFile');
const convertButton = document.getElementById('convertButton');
const buttonText = document.getElementById('buttonText');
const loading = document.getElementById('loading');
const errorInfo = document.getElementById('errorInfo');
const result = document.getElementById('result');
const resultContent = document.getElementById('resultContent');
const downloadButton = document.getElementById('downloadButton');

const go = new Go();

WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((result) => {
    go.run(result.instance);
    convertButton.disabled = false;
    buttonText.textContent = 'Convert Map';
    loading.classList.add('d-none');
}).catch((err) => {
    console.error(err);
    errorInfo.textContent = 'Failed to load converter. Please refresh the page.';
    errorInfo.classList.remove('d-none');
    buttonText.textContent = 'Unavailable';
    loading.classList.add('d-none');
});

const clearValidationErrors = () => {
    mapUrl.classList.remove('is-invalid');
    audioFile.classList.remove('is-invalid');
    mapFile.classList.remove('is-invalid');
};

form.addEventListener('submit', async (event) => {
    event.preventDefault();
    clearValidationErrors();

    const mapUrlValue = mapUrl.value.trim();
    const hasMapUrl = mapUrlValue !== '';
    const hasAudioFile = audioFile.files.length > 0;
    const hasMapFile = mapFile.files.length > 0;

    if (!hasMapUrl && !hasMapFile) {
        mapUrlFeedback.textContent = hasAudioFile
            ? 'Please provide a Sparebeat map URL or upload a local map file.'
            : 'Please provide a valid Sparebeat map URL.';
        mapUrl.classList.add('is-invalid');
        return;
    }

    if (hasMapFile && !hasAudioFile) {
        audioFile.classList.add('is-invalid');
        return;
    }

    try {
        setLoading(true);

        const audioFileData = audioFile.files[0];
        const mapFileData = mapFile.files[0];

        let mapId = null;
        let useBeta = false;

        if (!mapFileData) {
            const mapUrlMatch = mapUrlValue.match(SPAREBEAT_URL_REGEX);
            if (!mapUrlMatch) {
                mapUrlFeedback.textContent = 'Please provide a valid Sparebeat map URL.';
                mapUrl.classList.add('is-invalid');
                return;
            }
            useBeta = Boolean(mapUrlMatch[1]);
            mapId = mapUrlMatch[2];
        }

        buttonText.textContent = mapFileData ? 'Loading map...' : 'Downloading map...';
        const sbMap = await getSparebeatMap(mapId, mapFileData, useBeta);

        buttonText.textContent = 'Converting map...';
        const osuMap = convertSparebeatMap(JSON.stringify(sbMap));
        if (!osuMap.success) {
            throw new Error(osuMap.error || 'Unknown conversion error.');
        }

        buttonText.textContent = audioFileData ? 'Loading audio...' : 'Downloading audio...';
        const audioData = await getAudioData(mapId, audioFileData, useBeta);

        buttonText.textContent = 'Creating background...'
        const backgroundData = await createBackgroundImage(sbMap);

        buttonText.textContent = 'Creating .osz file...';
        const oszFile = await createOszFile(osuMap, audioData, backgroundData);

        const fileName = `${osuMap.metadata.artist} - ${osuMap.metadata.title}`;

        downloadButton.onclick = () => {
            const url = URL.createObjectURL(oszFile);
            const link = document.createElement('a');
            link.href = url;
            link.download = `${fileName}.osz`;
            link.click();
            URL.revokeObjectURL(url);
        };

        resultContent.innerHTML = `Successfully converted <strong>${fileName}</strong>`;
        result.classList.remove('d-none');
    } catch (err) {
        console.error(err);
        errorInfo.textContent = err.message;
        errorInfo.classList.remove('d-none');
    } finally {
        setLoading(false);
    }
});

mapUrl.addEventListener('input', clearValidationErrors);
audioFile.addEventListener('change', clearValidationErrors);
mapFile.addEventListener('change', clearValidationErrors);

const setLoading = (isLoading) => {
    loading.classList.toggle('d-none', !isLoading);
    convertButton.disabled = isLoading;

    if (isLoading) {
        errorInfo.classList.add('d-none');
        errorInfo.textContent = '';
        result.classList.add('d-none');
        resultContent.innerHTML = '';
    } else {
        buttonText.textContent = 'Convert Map';
    }
};

const fetchFromSparebeat = async (url, resourceType) => {
    try {
        const res = await fetch(url);
        if (res.status === 404) {
            throw new Error(`${resourceType} not found.`);
        }
        if (!res.ok) {
            throw new Error(`Failed to fetch ${resourceType.toLowerCase()}: ${res.status} ${res.statusText}`);
        }
        return res;
    } catch (err) {
        throw new Error(`Error fetching ${resourceType.toLowerCase()}: ${err.message}`);
    }
};

const getSparebeatMap = async (mapId, mapFile, useBeta) => {
    if (mapFile) {
        try {
            const map = await mapFile.text();
            return JSON.parse(map);
        } catch (err) {
            throw new Error(`Invalid map file: ${err.message}`);
        }
    }

    const mapURL = useBeta
        ? `${CORS_PROXY}https://beta.sparebeat.com/api/tracks/${mapId}/map`
        : `https://sparebeat.com/play/${mapId}/map`;

    const res = await fetchFromSparebeat(mapURL, 'Map');
    return await res.json();
};

const getAudioData = async (mapId, audioFile, useBeta) => {
    if (audioFile) {
        return new Uint8Array(await audioFile.arrayBuffer());
    }

    const audioURL = useBeta
        ? `${CORS_PROXY}https://beta.sparebeat.com/api/tracks/${mapId}/audio`
        : `https://sparebeat.com/play/${mapId}/music`;

    const res = await fetchFromSparebeat(audioURL, 'Audio');
    return new Uint8Array(await res.arrayBuffer());
};

const createBackgroundImage = async (sbMap) => {
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');
    canvas.width = 1920;
    canvas.height = 1080;

    const backgroundImg = new Image();
    backgroundImg.src = './images/background.png';
    await new Promise((resolve, reject) => {
        backgroundImg.onload = resolve;
        backgroundImg.onerror = () => reject(new Error('Error loading background image.'));
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
            reader.onload = () => resolve(new Uint8Array(reader.result));
            reader.readAsArrayBuffer(blob);
        }, 'image/png');
    });
};

const createOszFile = async (osuMap, audioData, backgroundData) => {
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
};