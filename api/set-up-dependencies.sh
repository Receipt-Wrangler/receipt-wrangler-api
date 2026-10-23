# Install build essential
apt-get install -y build-essential
# Install pip
apt-get update
apt-get install python3-pip -y
apt-get install python3.13-venv -y
# Set up venv
python3 -m venv wranglervenv
. wranglervenv/bin/activate
# Install pytorch cpu
pip3 install torch torchvision --index-url https://download.pytorch.org/whl/cpu
# Install easyocr
pip3 install easyocr
# Add lsb-release
apt-get update -y -qq
apt-get install apt-utils -y -qq
apt-get install lsb-release -y -qq
# Install dev files
apt-get install -y -qq libtesseract-dev libleptonica-dev
# Make sure english is installed
apt-get install -y -qq tesseract-ocr-eng

# For HEIC support
apt-get install -y -qq libde265-dev libheif-dev

# Install ImageMagick 7 with HEIC support
apt-get install pkg-config -y -qq
apt-get install -y -qq imagemagick-7.q16 libmagickwand-7.q16-dev

# Install Chromium for HTML-to-PDF rendering of email bodies via chromedp
apt-get install -y -qq chromium

# Adjust ImageMagick policy to allow for PDF conversion
POLICY_FILE="/etc/ImageMagick-7/policy.xml"
if [ -f "$POLICY_FILE" ]; then
    sed -i 's|<policy domain="coder" rights="none" pattern="PDF" />|<policy domain="coder" rights="read\|write" pattern="PDF" />|g' "$POLICY_FILE"
fi

# Verify HEIC support
magick -version | grep -i heic
magick -list format | grep -i heic
