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
apt-get install imagemagick libmagickwand-dev