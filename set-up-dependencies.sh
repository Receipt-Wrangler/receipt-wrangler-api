# Install pip
apt-get update
apt-get install python3-pip -y
apt-get install python3.11-venv -y

# Set up venv
python3 -m venv wranglervenv
source wranglervenv/bin/activate

# Install requirements
pip3 install -r ./imap-client/requirements.txt

# Install pytorch cpu
pip3 install torch torchvision --index-url https://download.pytorch.org/whl/cpu

# Install easyocr
pip3 install easyocr

# Add lsb-release
apt-get update -y -qq
apt-get install apt-utils -y -qq
apt-get install lsb-release -y -qq

# Install tesseract
apt-get install tesseract-ocr -y

# Install dev files
apt-get install -y -qq libtesseract-dev libleptonica-dev

# Make sure english is installed
apt-get install -y -qq tesseract-ocr-eng

# Install imageMagick
apt-get install -y -qq imagemagick libmagickwand-dev

# Adjust imageMagick policy to allow for pdf conversion
sed -i 's|<policy domain="coder" rights="none" pattern="PDF" />|<policy domain="coder" rights="read\|write" pattern="PDF" />|g' /etc/ImageMagick-6/policy.xml
