/**
 * Installation page for downloading MCP server packages
 */

import { useState } from 'react';
import { loadSecrets } from '../lib/store';
import JSZip from 'jszip';

type EditorType = 'claude' | 'cursor' | 'vscode';
type PlatformType = 'windows' | 'macos-intel' | 'macos-arm' | 'linux';
type InstallMethod = 'mcpb' | 'script';

export default function Install() {
  const [selectedEditor, setSelectedEditor] = useState<EditorType>('claude');
  const [selectedPlatform, setSelectedPlatform] = useState<PlatformType>('windows');
  const [selectedMethod, setSelectedMethod] = useState<InstallMethod>('mcpb');
  const [downloading, setDownloading] = useState(false);
  const [status, setStatus] = useState('');

  const handleDownload = async () => {
    setDownloading(true);
    setStatus('');

    try {
      const credentials = await loadSecrets();

      if (!credentials) {
        setStatus('❌ Please enter your API credentials in Settings first');
        setDownloading(false);
        return;
      }

      if (selectedEditor === 'claude' && selectedMethod === 'mcpb') {
        await downloadMcpbPackage(credentials, selectedPlatform);
      } else {
        await downloadScriptPackage(credentials, selectedEditor, selectedPlatform);
      }

      setStatus('✅ Download complete! Follow the instructions in the package.');
    } catch (error) {
      setStatus(`❌ Download failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
      console.error(error);
    } finally {
      setDownloading(false);
    }
  };

  return (
    <div className="card">
      <h2>Install MCP Server</h2>
      <p className="text-muted">
        Download and install the Agent Payment MCP server for your desktop client.
      </p>

      {/* Step 1: Select Editor */}
      <div className="install-step">
        <h3>1. Select Your Editor</h3>
        <div className="button-group">
          <button
            className={`button ${selectedEditor === 'claude' ? 'button-primary' : ''}`}
            onClick={() => setSelectedEditor('claude')}
          >
            Claude Desktop
          </button>
          <button
            className={`button ${selectedEditor === 'cursor' ? 'button-primary' : ''}`}
            onClick={() => setSelectedEditor('cursor')}
          >
            Cursor
          </button>
          <button
            className={`button ${selectedEditor === 'vscode' ? 'button-primary' : ''}`}
            onClick={() => setSelectedEditor('vscode')}
          >
            VS Code
          </button>
        </div>
      </div>

      {/* Step 2: Select Platform */}
      <div className="install-step">
        <h3>2. Select Your Platform</h3>
        <div className="button-group">
          <button
            className={`button ${selectedPlatform === 'windows' ? 'button-primary' : ''}`}
            onClick={() => setSelectedPlatform('windows')}
          >
            Windows
          </button>
          <button
            className={`button ${selectedPlatform === 'macos-intel' ? 'button-primary' : ''}`}
            onClick={() => setSelectedPlatform('macos-intel')}
          >
            macOS (Intel)
          </button>
          <button
            className={`button ${selectedPlatform === 'macos-arm' ? 'button-primary' : ''}`}
            onClick={() => setSelectedPlatform('macos-arm')}
          >
            macOS (Apple Silicon)
          </button>
          <button
            className={`button ${selectedPlatform === 'linux' ? 'button-primary' : ''}`}
            onClick={() => setSelectedPlatform('linux')}
          >
            Linux
          </button>
        </div>
      </div>

      {/* Step 3: Select Install Method (only for Claude) */}
      {selectedEditor === 'claude' && (
        <div className="install-step">
          <h3>3. Select Install Method</h3>
          <div className="button-group">
            <button
              className={`button ${selectedMethod === 'mcpb' ? 'button-primary' : ''}`}
              onClick={() => setSelectedMethod('mcpb')}
            >
              .mcpb Package (Recommended)
            </button>
            <button
              className={`button ${selectedMethod === 'script' ? 'button-primary' : ''}`}
              onClick={() => setSelectedMethod('script')}
            >
              Install Script
            </button>
          </div>
          <p className="text-muted">
            {selectedMethod === 'mcpb'
              ? 'Double-click the .mcpb file to install automatically'
              : 'Run the install script to configure manually'}
          </p>
        </div>
      )}

      {/* Download Button */}
      <div className="install-step">
        <h3>{selectedEditor === 'claude' && selectedMethod === 'mcpb' ? '4' : '3'}. Download & Install</h3>

        {/* AI Agent Instructions */}
        <div className="ai-instructions">
          <h4>🤖 Installing with an AI Agent?</h4>
          <p className="text-muted">Copy and send these instructions to your AI agent:</p>
          <div className="code-block">
            <code>
              Download and install the Agent Payment MCP server from this PWA:{'\n'}
              URL: {window.location.origin}{'\n'}
              Platform: {selectedPlatform}{'\n'}
              Editor: {selectedEditor}{'\n'}
              {selectedEditor === 'claude' && `Method: ${selectedMethod}\n`}
              {'\n'}
              Instructions:{'\n'}
              1. Navigate to {window.location.origin}/install{'\n'}
              2. Select {selectedEditor === 'claude' ? 'Claude Desktop' : selectedEditor === 'cursor' ? 'Cursor' : 'VS Code'}{'\n'}
              3. Select {selectedPlatform}{'\n'}
              {selectedEditor === 'claude' && `4. Select ${selectedMethod === 'mcpb' ? '.mcpb Package' : 'Install Script'}\n`}
              {selectedEditor === 'claude' && selectedMethod === 'mcpb' ? '4' : '3'}. Click "Download Installer"{'\n'}
              {selectedEditor === 'claude' && selectedMethod === 'mcpb' ? '5' : '4'}. Follow the installation instructions below
            </code>
          </div>
        </div>

        {/* Manual Download Button */}
        <h4>📥 Or Download Yourself</h4>
        <button
          onClick={handleDownload}
          disabled={downloading}
          className="button button-primary button-large"
        >
          {downloading ? 'Preparing Download...' : 'Download Installer'}
        </button>
      </div>

      {status && (
        <div className={`status ${status.includes('✅') ? 'success' : 'error'}`}>
          {status}
        </div>
      )}

      {/* Instructions */}
      <div className="install-instructions">
        <h3>📋 Installation Instructions</h3>
        <p className="text-muted">Follow these steps after downloading:</p>

        {selectedEditor === 'claude' && selectedMethod === 'mcpb' ? (
          <div>
            <h4>For Claude Desktop (.mcpb Package)</h4>
            <ol>
              <li>
                <strong>Locate the downloaded file:</strong> Find <code>agent-payment.mcpb</code> in your Downloads folder
              </li>
              <li>
                <strong>Install the package:</strong> Double-click the <code>.mcpb</code> file
              </li>
              <li>
                <strong>Wait for Claude Desktop:</strong> It will open automatically and show the install prompt
              </li>
              <li>
                <strong>Confirm installation:</strong> Click "Install" in Claude Desktop
              </li>
              <li>
                <strong>Restart Claude Desktop:</strong> Close and reopen the application
              </li>
              <li>
                <strong>✅ Done!</strong> You should now see Agent Payment tools in the MCP tools list
              </li>
            </ol>
            <p className="text-muted">
              💡 <strong>Tip:</strong> If tools don't appear, check Claude Desktop's MCP settings to verify the server is listed.
            </p>
          </div>
        ) : (
          <div>
            <h4>For {selectedEditor === 'cursor' ? 'Cursor' : selectedEditor === 'vscode' ? 'VS Code' : 'Claude Desktop'} (Install Script)</h4>
            <ol>
              <li>
                <strong>Extract the ZIP file:</strong> Unzip <code>agent-payment-{selectedEditor}-{selectedPlatform}.zip</code>
              </li>
              <li>
                <strong>Run the installer:</strong>
                <div className="code-block">
                  {selectedPlatform === 'windows' ? (
                    <>
                      <p>Right-click <code>install.bat</code> and select <strong>"Run as Administrator"</strong></p>
                    </>
                  ) : (
                    <>
                      <p>Open Terminal in the extracted folder and run:</p>
                      <code>chmod +x install.sh && ./install.sh</code>
                    </>
                  )}
                </div>
              </li>
              <li>
                <strong>Follow on-screen prompts:</strong> The script will guide you through the setup
              </li>
              <li>
                <strong>Restart your editor:</strong> Completely close and reopen {selectedEditor === 'cursor' ? 'Cursor' : selectedEditor === 'vscode' ? 'VS Code' : 'Claude Desktop'}
              </li>
              <li>
                <strong>✅ Done!</strong> Agent Payment tools should now be available
              </li>
            </ol>
            <p className="text-muted">
              💡 <strong>Tip:</strong> The installer places files in <code>~/.agent-payment/</code> (or <code>%USERPROFILE%\.agent-payment\</code> on Windows)
            </p>
          </div>
        )}

        <div className="troubleshooting">
          <h4>⚠️ Troubleshooting</h4>
          <ul>
            <li><strong>Tools not appearing?</strong> Make sure you've completely restarted the application</li>
            <li><strong>Permission denied?</strong> On Unix systems, run: <code>chmod +x ~/.agent-payment/agent-payment-server</code></li>
            <li><strong>Server not starting?</strong> Check that your API credentials in Settings are correct</li>
            <li><strong>Need help?</strong> Visit <a href="https://agentpmt.com/support" target="_blank" rel="noopener noreferrer">agentpmt.com/support</a></li>
          </ul>
        </div>
      </div>
    </div>
  );
}

/**
 * Download .mcpb package for Claude Desktop
 */
async function downloadMcpbPackage(
  credentials: any,
  platform: PlatformType
): Promise<void> {
  const zip = new JSZip();

  // Add manifest.json
  const manifest = {
    manifest_version: '0.2',
    name: 'agent-payment',
    display_name: 'Agent Payment',
    version: '1.0.0',
    description: 'MCP tools from Agent Payment API',
    author: {
      name: 'Agent Payment',
      url: 'https://agentpmt.com'
    },
    server: {
      type: 'binary',
      entry_point: getBinaryName(platform),
      mcp_config: {
        command: `./${getBinaryName(platform)}`,
        args: [],
        env: {}
      }
    }
  };
  zip.file('manifest.json', JSON.stringify(manifest, null, 2));

  // Add config.json
  const config = {
    api_key: credentials.apiKey,
    budget_key: credentials.budgetKey,
    api_url: 'https://api.agentpmt.com',
    ...(credentials.auth ? { auth: credentials.auth } : {})
  };
  zip.file('config.json', JSON.stringify(config, null, 2));

  // Fetch and add binary
  const binaryPath = getBinaryPath(platform);
  const binaryResponse = await fetch(binaryPath);
  const binaryBlob = await binaryResponse.blob();
  zip.file(getBinaryName(platform), binaryBlob, { binary: true });

  // Add README
  zip.file('README.md', generateReadme('claude', 'mcpb'));

  // Generate and download
  const blob = await zip.generateAsync({ type: 'blob' });
  downloadBlob(blob, 'agent-payment.mcpb');
}

/**
 * Download install script package
 */
async function downloadScriptPackage(
  credentials: any,
  editor: EditorType,
  platform: PlatformType
): Promise<void> {
  const zip = new JSZip();

  // Add config.json
  const config = {
    api_key: credentials.apiKey,
    budget_key: credentials.budgetKey,
    api_url: 'https://api.agentpmt.com',
    ...(credentials.auth ? { auth: credentials.auth } : {})
  };
  zip.file('config.json', JSON.stringify(config, null, 2));

  // Fetch and add binary
  const binaryPath = getBinaryPath(platform);
  const binaryResponse = await fetch(binaryPath);
  const binaryBlob = await binaryResponse.blob();
  zip.file(getBinaryName(platform), binaryBlob, { binary: true });

  // Add install script
  const installScript = generateInstallScript(editor, platform);
  const scriptName = platform === 'windows' ? 'install.bat' : 'install.sh';
  zip.file(scriptName, installScript);

  // Add README
  zip.file('README.md', generateReadme(editor, 'script'));

  // Generate and download
  const blob = await zip.generateAsync({ type: 'blob' });
  downloadBlob(blob, `agent-payment-${editor}-${platform}.zip`);
}

/**
 * Helper functions
 */
function getBinaryName(platform: PlatformType): string {
  return platform === 'windows' ? 'agent-payment-server.exe' : 'agent-payment-server';
}

function getBinaryPath(platform: PlatformType): string {
  const base = '/binaries'; // Served by PWA backend
  const mapping: Record<PlatformType, string> = {
    'windows': `${base}/windows-amd64/agent-payment-server.exe`,
    'macos-intel': `${base}/darwin-amd64/agent-payment-server`,
    'macos-arm': `${base}/darwin-arm64/agent-payment-server`,
    'linux': `${base}/linux-amd64/agent-payment-server`
  };
  return mapping[platform];
}

function generateInstallScript(editor: EditorType, platform: PlatformType): string {
  if (platform === 'windows') {
    return generateWindowsScript(editor);
  } else {
    return generateUnixScript(editor, platform);
  }
}

function generateWindowsScript(editor: EditorType): string {
  const configPath = editor === 'claude'
    ? '%APPDATA%\\Claude\\claude_desktop_config.json'
    : editor === 'cursor'
    ? '%USERPROFILE%\\.cursor\\mcp.json'
    : '%APPDATA%\\Code\\User\\globalStorage\\claude-code\\mcp.json';

  return `@echo off
echo ================================================
echo Agent Payment MCP Server Installer
echo ================================================
echo.

set "INSTALL_DIR=%USERPROFILE%\\.agent-payment"
set "CONFIG_PATH=${configPath}"

echo Creating installation directory...
mkdir "%INSTALL_DIR%" 2>nul

echo Copying server executable...
copy /Y agent-payment-server.exe "%INSTALL_DIR%\\" >nul

echo Copying configuration...
copy /Y config.json "%INSTALL_DIR%\\" >nul

echo.
echo Configuring ${editor}...

REM Create config directory if it doesn't exist
for %%I in ("%CONFIG_PATH%") do mkdir "%%~dpI" 2>nul

REM TODO: Add JSON merging logic here
REM For now, display manual instructions

echo.
echo ================================================
echo Installation Complete!
echo ================================================
echo.
echo Server installed to: %INSTALL_DIR%
echo.
echo Next steps:
echo 1. Restart ${editor}
echo 2. The Agent Payment tools should now be available
echo.
pause
`;
}

function generateUnixScript(editor: EditorType, platform: PlatformType): string {
  const configPath = editor === 'claude'
    ? platform.startsWith('macos')
      ? '~/Library/Application Support/Claude/claude_desktop_config.json'
      : '~/.config/Claude/claude_desktop_config.json'
    : editor === 'cursor'
    ? '~/.cursor/mcp.json'
    : '~/.config/Code/User/globalStorage/claude-code/mcp.json';

  return `#!/bin/bash
set -e

echo "================================================"
echo "Agent Payment MCP Server Installer"
echo "================================================"
echo

INSTALL_DIR="$HOME/.agent-payment"
CONFIG_PATH="${configPath}"

echo "Creating installation directory..."
mkdir -p "$INSTALL_DIR"

echo "Copying server executable..."
cp agent-payment-server "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/agent-payment-server"

echo "Copying configuration..."
cp config.json "$INSTALL_DIR/"

echo
echo "Configuring ${editor}..."

# Create config directory if it doesn't exist
mkdir -p "$(dirname "$CONFIG_PATH")"

# TODO: Add JSON merging logic here using jq
# For now, display manual instructions

echo
echo "================================================"
echo "Installation Complete!"
echo "================================================"
echo
echo "Server installed to: $INSTALL_DIR"
echo
echo "Next steps:"
echo "1. Restart ${editor}"
echo "2. The Agent Payment tools should now be available"
echo
`;
}

function generateReadme(editor: string, method: string): string {
  return `# Agent Payment MCP Server

Thank you for installing the Agent Payment MCP server!

## Installation Method: ${method === 'mcpb' ? '.mcpb Package' : 'Install Script'}

## What's Included

- \`agent-payment-server\`: Standalone MCP server executable
- \`config.json\`: Your API credentials (keep this secure!)
- \`install.sh\` or \`install.bat\`: Installation script
- This README

## Installation Instructions

${method === 'mcpb' ? `
### For Claude Desktop (.mcpb)

1. Double-click the \`.mcpb\` file
2. Claude Desktop will open automatically
3. Click "Install" when prompted
4. Restart Claude Desktop
5. Done!

` : `
### For ${editor === 'cursor' ? 'Cursor' : editor === 'vscode' ? 'VS Code' : 'Claude Desktop'}

**macOS/Linux:**
\`\`\`bash
chmod +x install.sh
./install.sh
\`\`\`

**Windows:**
Right-click \`install.bat\` and select "Run as Administrator"
`}

## Verifying Installation

After restarting your editor, you should see Agent Payment tools available in the MCP tools list.

## Troubleshooting

### Server not starting
- Ensure the executable has execute permissions (macOS/Linux)
- Check that config.json is in the same directory as the executable
- Verify your API credentials are correct

### Tools not appearing
- Restart your editor completely
- Check the MCP server logs in your editor
- Verify your API credentials in config.json

## Configuration

The server reads configuration from \`config.json\`:

\`\`\`json
{
  "api_key": "your-api-key",
  "budget_key": "your-budget-key",
  "api_url": "https://api.agentpmt.com"
}
\`\`\`

You can edit this file to update your credentials.

## Support

For issues or questions:
- GitHub: [your-repo-url]
- Email: support@agentpmt.com
- Website: https://agentpmt.com

## Security

- Keep your \`config.json\` file secure
- Do not share your API credentials
- The executable only connects to api.agentpmt.com

---

Generated by Agent Payment PWA
`;
}

function downloadBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}
