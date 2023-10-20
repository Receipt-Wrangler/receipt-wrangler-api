
# Adjust imageMagick policy to allow for pdf conversion
sed -i 's|<policy domain="coder" rights="none" pattern="PDF" />|<policy domain="coder" rights="read\|write" pattern="PDF" />|g' etc/ImageMagick-6/policy.xml

# Start api
./api --env prod