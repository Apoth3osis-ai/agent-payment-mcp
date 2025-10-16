#!/usr/bin/env python3
"""Validate server.json structure based on MCP registry documentation."""

import json
import sys

def validate_server_json(config):
    """Validate the server.json configuration."""
    errors = []
    warnings = []

    # Required fields
    required_fields = ['name', 'version', 'description']
    for field in required_fields:
        if field not in config:
            errors.append(f"Missing required field: {field}")
        elif not config[field]:
            errors.append(f"Field '{field}' cannot be empty")

    # Validate name format (should be namespace/server-name)
    if 'name' in config:
        name = config['name']
        if not name.startswith('io.github.') and not name.startswith('com.'):
            warnings.append(f"Name should use io.github.* or com.* namespace: {name}")
        if '/' not in name:
            warnings.append(f"Name should include a slash separator (namespace/name): {name}")

    # Check for either packages or remotes
    has_packages = config.get('packages') and len(config['packages']) > 0
    has_remotes = config.get('remotes') and len(config['remotes']) > 0

    if not has_packages and not has_remotes:
        errors.append("Must have at least one package or remote deployment")

    # Validate packages if present
    if has_packages:
        for idx, package in enumerate(config['packages']):
            pkg_type = package.get('type')
            if not pkg_type:
                errors.append(f"Package {idx}: Missing 'type' field")
            elif pkg_type not in ['npm', 'pypi', 'mcpb', 'docker', 'nuget']:
                warnings.append(f"Package {idx}: Unknown package type '{pkg_type}'")

            # Validate MCPB packages
            if pkg_type == 'mcpb':
                if 'uri' not in package:
                    errors.append(f"Package {idx}: MCPB package missing 'uri' field")
                if 'sha256' not in package:
                    errors.append(f"Package {idx}: MCPB package missing 'sha256' field")
                if 'platforms' not in package or not package['platforms']:
                    errors.append(f"Package {idx}: MCPB package missing 'platforms' array")
                else:
                    # Validate platforms
                    for pidx, platform in enumerate(package['platforms']):
                        if 'os' not in platform:
                            errors.append(f"Package {idx}, Platform {pidx}: Missing 'os' field")
                        if 'arch' not in platform:
                            errors.append(f"Package {idx}, Platform {pidx}: Missing 'arch' field")
                        if 'uri' not in platform:
                            errors.append(f"Package {idx}, Platform {pidx}: Missing 'uri' field")
                        if 'sha256' not in platform:
                            errors.append(f"Package {idx}, Platform {pidx}: Missing 'sha256' field")

    # Validate remotes if present
    if has_remotes:
        for idx, remote in enumerate(config['remotes']):
            if 'endpoint' not in remote:
                errors.append(f"Remote {idx}: Missing 'endpoint' field")
            if 'transport' not in remote:
                errors.append(f"Remote {idx}: Missing 'transport' field")
            elif remote['transport'] not in ['streamable-http', 'sse']:
                warnings.append(f"Remote {idx}: Unknown transport '{remote['transport']}'")

    # Recommended fields
    recommended_fields = ['license', 'homepage', 'repository']
    for field in recommended_fields:
        if field not in config:
            warnings.append(f"Recommended field missing: {field}")

    # Validate repository structure if present
    if 'repository' in config:
        repo = config['repository']
        if not isinstance(repo, dict):
            errors.append("Repository must be an object")
        else:
            if 'type' not in repo:
                warnings.append("Repository missing 'type' field")
            if 'url' not in repo:
                warnings.append("Repository missing 'url' field")

    return errors, warnings

def main():
    # Read the server.json file
    try:
        with open('server.json', 'r') as f:
            config = json.load(f)
    except FileNotFoundError:
        print("❌ Error: server.json not found")
        sys.exit(1)
    except json.JSONDecodeError as e:
        print(f"❌ Error: Invalid JSON in server.json: {e}")
        sys.exit(1)

    print("Validating server.json...\n")

    # Validate the configuration
    errors, warnings = validate_server_json(config)

    # Display results
    if errors:
        print("❌ ERRORS:")
        for error in errors:
            print(f"  • {error}")
        print()

    if warnings:
        print("⚠️  WARNINGS:")
        for warning in warnings:
            print(f"  • {warning}")
        print()

    if not errors and not warnings:
        print("✅ server.json looks good!")
    elif not errors:
        print("✅ server.json is valid (with warnings above)")
    else:
        print("❌ server.json has validation errors")
        sys.exit(1)

    # Display summary
    print("\n📋 Configuration Summary:")
    print(f"  Name: {config.get('name')}")
    print(f"  Version: {config.get('version')}")
    print(f"  Description: {config.get('description', '')[:60]}...")
    print(f"  License: {config.get('license', 'Not specified')}")
    print(f"  Homepage: {config.get('homepage', 'Not specified')}")
    print(f"  Packages: {len(config.get('packages', []))}")
    print(f"  Remotes: {len(config.get('remotes', []))}")

    # Show platform support
    if config.get('packages'):
        for pkg in config.get('packages', []):
            if pkg.get('type') == 'mcpb' and pkg.get('platforms'):
                print(f"\n  🖥️  Platform Support ({len(pkg['platforms'])} platforms):")
                for platform in pkg['platforms']:
                    os_name = platform.get('os', 'unknown')
                    arch = platform.get('arch', 'unknown')
                    print(f"    • {os_name}/{arch}")

    return 0

if __name__ == '__main__':
    sys.exit(main())
