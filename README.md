# Tor Droplet

## Overview

**Tor Droplet** is a lightweight tool to quickly set up a Tor proxy on a DigitalOcean Droplet. It automates the creation and configuration of a virtual machine, sets up a Tor relay, and provides a local proxy interface for secure browsing and data routing.

This project is ideal for developers and enthusiasts who need a quick and disposable Tor proxy for testing, privacy, or secure routing purposes.

## Features
- **Automated Droplet Setup**: Provision and configure a DigitalOcean Droplet with a single command.
- **Customizable Tor Configuration**: Specify Tor exit nodes, DNS settings, and other options via YAML.
- **Local Proxy Support**: Provides a local TCP proxy interface for routing traffic through the remote Tor relay.
- **TTL Management**: Automatically destroys the Droplet after a specified time-to-live (TTL).
- **List DigitalOcean Resources**: View available images, regions, and droplet sizes.
- **Destroy Expired Droplets**: Clean up lingering Droplets by using the `destroy-expired` action to avoid unnecessary costs.

## Prerequisites

- A DigitalOcean account and API token with appropriate permissions:
    - droplets: create, read, update, delete
    - regions: read
    - sizes: read
    - image: read
- Go (Golang) installed on your system.

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/toxyl/tor-droplet.git
   cd tor-droplet
   ```

2. Install dependencies:  
   Ensure you have the necessary Go modules installed by running:
   ```bash
   go mod tidy
   ```

3. Create a configuration file:  
   Copy the example configuration file and edit it with your details:
   ```bash
   cp example-config.yaml config.yaml
   ```

4. Update the API token and other settings in `config.yaml`.

## Usage

Run the tool with the following commands:

### Basic Commands

- **Create a Tor proxy droplet**:
  ```bash
  go run . config.yaml
  ```

- **List available images**:
  ```bash
  go run . config.yaml list-images
  ```

- **List available regions**:
  ```bash
  go run . config.yaml list-regions
  ```

- **List available droplet sizes**:
  ```bash
  go run . config.yaml list-sizes
  ```

- **Destroy expired droplets**:
  ```bash
  go run . config.yaml destroy-expired
  ```

### Cleanup

The Droplet is automatically deleted after the configured TTL. However, if the application or the host crash, the droplet will keep running leading to unnecessary costs. The tool will attempt to remove expired instances before creating a new instance, after the TTL of the current instance has expired and if the user exits the application with CTRL+C. 

## Configuration

The `config.yaml` file allows you to customize the behavior of the Tor Droplet. Below is an example:

```yaml
api_token: your_api_token_here
droplet:
  size: s-1vcpu-512mb-10gb
  region: nyc1
  image: ubuntu-24-10-x64
  ttl: 10m
tor_options:
  - "ExitNodes {us},{ca} STRICT"
ports:
  local: 8888
  remote: 9050
dns:
  primary: 8.8.8.8
  secondary: 8.8.4.4
```

### Key Configuration Options

- `api_token`: Your DigitalOcean API token.
- `droplet.size`: Size of the droplet (use `list-sizes` to find available options).
- `droplet.region`: Region of the droplet (use `list-regions` to find available options).
- `droplet.image`: Base image for the droplet (use `list-images` to find available options).
- `droplet.ttl`: Time-to-live for the droplet before automatic cleanup.
- `tor_options`: Tor configuration options such as specific exit nodes.
- `ports.local`: Local proxy port.
- `ports.remote`: Remote Tor service port.
- `dns`: Primary and secondary DNS servers for the droplet.

## Important Notes

1. **Traffic Encryption**: The traffic between the local port and the remote port is **not encrypted** by default. If you are using HTTP traffic, it will be unencrypted, similar to using any standard browser. Ensure that you use HTTPS or other secure methods to protect sensitive data.

2. **Droplet Cleanup**: Ensure that you regularly run the `destroy-expired` action or monitor your droplets manually. The tool includes mechanisms to identify and remove expired instances automatically but this is a proof of concept, so there might be errors leaving instances running - check manually every once in a while to avoid exploding costs.

## License

This project is licensed under [The Unlicense](https://unlicense.org). You are free to use, modify, and distribute this software without restriction.

## Disclaimer

This tool is provided as-is, without warranty of any kind. Use at your own risk. Ensure compliance with local laws and DigitalOcean's terms of service when using this tool.
