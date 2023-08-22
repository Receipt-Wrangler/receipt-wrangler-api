# Add lsb-release
sudo apt-get update -y -qq
sudo apt-get install apt-utils -y -qq
sudo apt-get install lsb-release -y -qq

# Add repo source
echo "deb https://notesalexp.org/tesseract-ocr5/$(lsb_release -cs)/ $(lsb_release -cs) main" \
| sudo tee /etc/apt/sources.list.d/notesalexp.list > /dev/null

# Add repo key
sudo apt-get update  -y -oAcquire::AllowInsecureRepositories=true --allow-unauthenticated
sudo apt-get install -y notesalexp-keyring -oAcquire::AllowInsecureRepositories=true --allow-unauthenticated
sudo apt-get update -y --allow-unauthenticated

# Install tesseract
sudo apt-get install tesseract-ocr -y

# Install dev files
sudo apt-get install -y -qq libtesseract-dev libleptonica-dev

# Make sure english is installed
sudo apt-get install -y -qq tesseract-ocr-eng

# Install imageMagick
sudo apt-get install -y -qq imagemagick libmagickwand-dev