openssl genrsa -out rsa_private_key.pem 4096 
openssl rsa -in rsa_private_key.pem -pubout -out rsa_public_key.pem