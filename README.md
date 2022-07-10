# groundcover CLI

## Installation

### From Script

CLI now has an installer script that will automatically grab the latest version and install it locally.

`sh -c "$(curl -fsSL https://groundcover.com/install.sh)"`

### From the Binary Releases

Binary downloads of the CLI can be found on [the Releases page](https://github.com/groundcover-com/cli/releases/latest).

#### Linux

```bash
VERSION=0.1.0

# Intel Chip
curl -SsL https://github.com/groundcover-com/cli/releases/download/v${VERSION}/groundcover_${VERSION}_linux_amd64.tar.gz -o /tmp/groundcover.tar.gz
# ARM chip
curl -SsL https://github.com/groundcover-com/cli/releases/download/v${VERSION}/groundcover_${VERSION}_linux_arm64.tar.gz -o /tmp/groundcover.tar.gz

mkdir -p ~/.groundcover/bin
tar -zxf /tmp/groundcover.tar.gz -C ~/.groundcover/bin
chmod +x ~/.groundcover/bin/groundcover

echo 'export PATH=~/.groundcover/bin:/$PATH' >> ~/.bashrc
```

#### MacOS

```bash
VERSION=0.1.0

# Intel chip
curl -SsL https://github.com/groundcover-com/cli/releases/download/v${VERSION}/groundcover_${VERSION}_darwin_amd64.tar.gz -o /tmp/groundcover.tar.gz
# Apple chip
curl -SsL https://github.com/groundcover-com/cli/releases/download/v${VERSION}/groundcover_${VERSION}_darwin_arm64.tar.gz -o /tmp/groundcover.tar.gz

mkdir -p ~/.groundcover/bin
tar -zxf /tmp/groundcover.tar.gz -C ~/.groundcover/bin
chmod +x ~/.groundcover/bin/groundcover

echo 'export PATH=~/.groundcover/bin:/$PATH' >> ~/.zshrc
```
