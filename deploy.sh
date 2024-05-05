#!/bin/bash

git pull
# Function to check if a command is available
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check and install Homebrew if not available
if ! command_exists brew; then
    echo "Homebrew is not installed. Installing..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
fi

# Check and install Go if not available
if ! command_exists go; then
    echo "Go is not installed. Installing..."
    if command_exists brew; then
        brew install go
    elif command_exists apt-get; then
        sudo apt-get install -y golang
    elif command_exists yum; then
        sudo yum install -y golang
    else
        echo "Unsupported package manager. Please install Go manually."
        exit 1
    fi
fi

# Check and install Mage if not available
if ! command_exists mage; then
    echo "Mage is not installed. Installing..."
    go install github.com/magefile/mage@latest
fi

# Check and install npm if not available
if ! command_exists npm; then
    echo "npm is not installed. Installing..."
    if command_exists brew; then
        brew install npm
    elif command_exists apt-get; then
        sudo apt-get install -y npm
    elif command_exists yum; then
        sudo yum install -y npm
    else
        echo "Unsupported package manager. Please install npm manually."
        exit 1
    fi
fi

# Execute mage command
mage -v

npm install

# Execute npm run build
npm run build

# Check and install Docker Compose if not available
if ! command_exists docker-compose; then
    echo "Docker Compose is not installed. Installing..."
    if command_exists brew; then
        brew install docker-compose
    elif command_exists apt-get; then
        sudo apt-get install -y docker-compose
    elif command_exists yum; then
        sudo yum install -y docker-compose
    else
        echo "Unsupported package manager. Please install Docker Compose manually."
        exit 1
    fi
fi

# Export GRAFANA_ACCESS_POLICY_TOKEN
export GRAFANA_ACCESS_POLICY_TOKEN=glc_eyJvIjoiNjQyNjM1IiwibiI6InBsdWdpbi1zaWduaW5nLXBsdWdpbi1zaWduaW5nLXRva2VuIiwiayI6IkkyNXk1QklOVUM4MTVEOE5RejNHR3g5OCIsIm0iOnsiciI6InVzIn19

# Execute npx @grafana/sign-plugin
npx @grafana/sign-plugin@latest --rootUrls http://localhost:3000

# Run Docker Compose
docker-compose up -d && docker compose restart