# Add python dependencies
apt-get update
apt-get install python3.9 -y
apt-get install python3.9-distutils  -y

# Install pip
wget https://bootstrap.pypa.io/get-pip.py
python3 get-pip.py
rm get-pip.py

pip3 install -r ./imap-client/requirements.txt

# Install pytorch cpu
pip3 install torch torchvision --index-url https://download.pytorch.org/whl/cpu

# Install easyocr
pip3 install easyocr

# Add lsb-release
apt-get update -y -qq
apt-get install apt-utils -y -qq
apt-get install lsb-release -y -qq

# Add repo source
echo "deb https://notesalexp.org/tesseract-ocr5/$(lsb_release -cs)/ $(lsb_release -cs) main" \
| tee /etc/apt/sources.list.d/notesalexp.list > /dev/null

# Add repo key
apt-get update  -y -oAcquire::AllowInsecureRepositories=true --allow-unauthenticated
apt-get install -y notesalexp-keyring -oAcquire::AllowInsecureRepositories=true --allow-unauthenticated
apt-get update -y --allow-unauthenticated

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
