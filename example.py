import base64
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.asymmetric.padding import MGF1, OAEP
from cryptography.hazmat.primitives.asymmetric.rsa import RSAPrivateKey
from cryptography.hazmat.primitives.serialization import load_pem_private_key


PRIVATE_KEY = open('./rsa_private_key.pem', 'r').read()
private_key_bytes = PRIVATE_KEY.encode("utf-8")
private_key: RSAPrivateKey = load_pem_private_key(private_key_bytes, None)

def decode_message(data: str) -> str:
    padding = OAEP(mgf=MGF1(algorithm=hashes.SHA256()), algorithm=hashes.SHA256(), label=None)
    decrypted_message = private_key.decrypt(base64.b64decode(data.encode("utf-8")), padding)
    return decrypted_message.decode('utf-8')


print(decode_message('H1zeIJMINkKaoyHac5EAD3lHFd+ySDbwRhnYvo3BjIDoA3SmNEYx9rrlqh2i72jaFaZfK0Isk8ojxBwrO+r6x8G886m6aFfatulbHluumSNzEN8QFX8Dm2LKG4Bb25eG+jks++SKEysvo/OwHaVZ5Wk8DaNZdfzneOtmnq+0lAM='))