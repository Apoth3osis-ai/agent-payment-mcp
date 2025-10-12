let tools = [];
let selectedTools = [];

// Load tools on page load
window.addEventListener('DOMContentLoaded', async () => {
    try {
        const response = await fetch('/api/detect');
        tools = await response.json();

        renderTools();

        // Auto-select detected tools
        selectedTools = tools.filter(t => t.detected).map(t => t.id);
        updateCheckboxes();
    } catch (err) {
        showError(`Failed to detect tools: ${err.message}`);
    }
});

// Render tools list
function renderTools() {
    const toolsList = document.getElementById('tools-list');
    toolsList.innerHTML = '';

    tools.forEach(tool => {
        const label = document.createElement('label');
        label.className = 'tool-checkbox';

        const checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.dataset.toolId = tool.id;
        checkbox.addEventListener('change', (e) => {
            if (e.target.checked) {
                if (!selectedTools.includes(tool.id)) {
                    selectedTools.push(tool.id);
                }
            } else {
                selectedTools = selectedTools.filter(id => id !== tool.id);
            }
        });

        const nameSpan = document.createElement('span');
        nameSpan.className = 'tool-name';
        nameSpan.textContent = tool.name;

        if (tool.detected) {
            const badge = document.createElement('span');
            badge.className = 'detected-badge';
            badge.textContent = '✓ Detected';
            nameSpan.appendChild(badge);
        }

        label.appendChild(checkbox);
        label.appendChild(nameSpan);

        if (tool.configPath && tool.configPath !== 'GUI Configuration Required') {
            const pathSpan = document.createElement('span');
            pathSpan.className = 'config-path';
            pathSpan.textContent = tool.configPath;
            label.appendChild(pathSpan);
        }

        toolsList.appendChild(label);
    });
}

// Update checkboxes based on selectedTools
function updateCheckboxes() {
    document.querySelectorAll('.tool-checkbox input[type="checkbox"]').forEach(checkbox => {
        const toolId = checkbox.dataset.toolId;
        checkbox.checked = selectedTools.includes(toolId);
    });
}

// Install button handler
document.getElementById('install-btn').addEventListener('click', async () => {
    const apiKey = document.getElementById('apiKey').value.trim();
    const budgetKey = document.getElementById('budgetKey').value.trim();
    const environment = document.getElementById('environment').value;

    // Validate
    if (!apiKey) {
        showError('Please enter your API Key');
        return;
    }

    if (!budgetKey) {
        showError('Please enter your Budget Key');
        return;
    }

    if (selectedTools.length === 0) {
        showError('Please select at least one AI tool');
        return;
    }

    hideError();

    // Disable form
    document.getElementById('install-btn').disabled = true;
    document.querySelectorAll('input, select').forEach(el => el.disabled = true);

    // Show progress
    document.getElementById('progress-section').style.display = 'block';

    // Call install API
    try {
        const response = await fetch('/api/install', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                apiKey,
                budgetKey,
                environment,
                selectedTools
            })
        });

        const result = await response.json();

        if (result.error) {
            showError(result.error);
            document.getElementById('install-btn').disabled = false;
            document.querySelectorAll('input, select').forEach(el => el.disabled = false);
            document.getElementById('progress-section').style.display = 'none';
        } else {
            // Update progress
            document.getElementById('progress-message').textContent = result.message;
            document.getElementById('progress-fill').style.width = result.progress + '%';
            document.getElementById('progress-text').textContent = result.progress + '%';

            if (result.step === 'complete') {
                document.getElementById('form-container').style.display = 'none';
                document.getElementById('progress-section').style.display = 'none';
                document.getElementById('success-box').style.display = 'block';

                // Show Windows Defender warning on Windows
                if (navigator.platform.toLowerCase().includes('win')) {
                    document.getElementById('windows-defender-warning').style.display = 'block';
                }
            }
        }
    } catch (err) {
        showError(`Installation failed: ${err.message}`);
        document.getElementById('install-btn').disabled = false;
        document.querySelectorAll('input, select').forEach(el => el.disabled = false);
        document.getElementById('progress-section').style.display = 'none';
    }
});

function showError(message) {
    document.getElementById('error-message').textContent = message;
    document.getElementById('error-box').style.display = 'block';
}

function hideError() {
    document.getElementById('error-box').style.display = 'none';
}
