# Install build essential
apt-get install -y build-essential
# Install pip
apt-get update
apt-get install python3-pip -y
apt-get install python3.13-venv -y
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
# Install dev files
apt-get install -y -qq libtesseract-dev libleptonica-dev
# Make sure english is installed
apt-get install -y -qq tesseract-ocr-eng

# For HEIC support
apt-get install -y -qq libde265-dev libheif-dev

# Build ImageMagick from source
apt-get install git -y -qq
apt-get install pkg-config -y -qq
apt-get install -y -qq libmagickwand-dev
# Clone ImageMagick
git clone --depth 1 https://github.com/ImageMagick/ImageMagick.git ImageMagick-7.1.1
# Build ImageMagick
cd ImageMagick-7.1.1
# Configure with HEIC support explicitly enabled
./configure --with-heic=yes
make VERBOSE=1
make install
ldconfig
# Cleanup
cd ..
rm -rf ImageMagick-7.1.1
# Adjust imageMagick policy to allow for pdf conversion
sed -i 's|<policy domain="coder" rights="none" pattern="PDF" />|<policy domain="coder" rights="read\|write" pattern="PDF" />|g' /usr/local/etc/ImageMagick-7/policy.xml

# Verify HEIC support
magick -version | grep -i heic
magick -list format | grep -i heic
