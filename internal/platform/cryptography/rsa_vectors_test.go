package cryptography

import (
	"crypto/rsa"
	"encoding/hex"
	"math/big"
	"testing"
)

// Independent Wycheproof 0.9 signature vectors, tcIds 1, 62 and 108.
// Source: https://raw.githubusercontent.com/C2SP/wycheproof/main/testvectors_v1/rsa_pss_2048_sha256_mgf1_32_test.json
// Source SHA-256: 7f6efafc160f4816b96cbf1c12188a31051d7e3f001e27505d9edb5f2a0e325c
// Public test data only; tests do not download or regenerate these signatures.
func TestRSAIndependentVectors(t *testing.T) {
	modulus, ok := new(big.Int).SetString("00a2b451a07d0aa5f96e455671513550514a8a5b462ebef717094fa1fee82224e637f9746d3f7cafd31878d80325b6ef5a1700f65903b469429e89d6eac8845097b5ab393189db92512ed8a7711a1253facd20f79c15e8247f3d3e42e46e48c98e254a2fe9765313a03eff8f17e1a029397a1fa26a8dce26f490ed81299615d9814c22da610428e09c7d9658594266f5c021d0fceca08d945a12be82de4d1ece6b4c03145b5d3495d4ed5411eb878daf05fd7afc3e09ada0f1126422f590975a1969816f48698bcbba1b4d9cae79d460d8f9f85e7975005d9bc22c4e5ac0f7c1a45d12569a62807d3b9a02e5a530e773066f453d1f5b4c2e9cf7820283f742b9d5", 16)
	if !ok {
		t.Fatal("invalid vector modulus")
	}
	key := &rsa.PublicKey{N: modulus, E: 65537}
	for _, vector := range []struct {
		id                 string
		message, signature string
		pss, pkcs          bool
	}{
		{"1", "", "4f01e0c12b08625ecac89a69231906edf826380f37c959a96690d046316d68ffce9d5c471694fcebfc6b45534864689256e4fc81c78e583f675d0c94b449647451e81beff01a11a516d5e5ce3f1a910437cb8a3a5096b19fb15f4524a35b23d89cdba12cf5b71aac1047b28c562df7c5542c34ce23a182cf7e0e231934b17294799d44877a1d68ef1b8f073619b7618e6b7c22db20030d98cf591ffc3d4da5f58613ecd5ecfc3b40a1d02f40891ca43695cd4c088b05a8054c89c595a47e274816f35384226f74459ee63e25a1bfc03c360490552ec38343f8ace502f065303b00bc0ec320711b211fde92e57feb9013c3609342495ec0d7cabdec21e54acc38", true, false},
		{"62", "313233343030", "67d1d1c0a398148625317c3f5e44b738bdf461c27a59594b39ebb2aebef233c7809379e54411411b82d2e7ac88f989b58373d532c758baea121878ce9759441738d121881c1fa2d04421f02dd565b12770d844611ed1873a0b64d822709a6b78d6d3892b294404bce6711001d6c3a54546c76a1d17819674b0be904497a233b466fe4becc832dee740f9ab79e5b9f5db0b0f9aac0084ba05cebf42303b5ca2ad95e3d61b29ed6475545c02e93e7b0e118af92f5cddb1faeb2cbc23c9e69c120e29df7fe31991e887b3b29e77688c60e80be65cccf3d7861a7a14c39e6a6e5645568e2cc5e4a17b75db1dd415aadb45e112a9b582b2ff6e82a43d7a7347b7b56d", false, false},
		{"108", "313233343030", "1758eb94588e6fc4f50c1be1afcaa41027869f304cad513b1fb12c2f446d63cdc05c4830a7e3e630da7b2da4f7867cc173bf6420f9732277282596de41ded32e21d0cc31441174da8765f57419c7764ea758f55bc17646eb100c435d1ac0eed6fc7ba6de5f832094ee2f479979765e05ac9976788db3c241a9e32a0da864f0019a87646ba623d63f4411af5dee1be9ec488c7e3e1b231479de70b9ac5f78a17b1f4120aece45f26c07e7bb345fdfeb05e14bcaacc614672a465fc523624cb19f66f9c6c3f642b832ca44cb25176d679f0e05606c3fed022cac24c2bf960a406d48818e3eb7ed53b0446032469047dfed95fc18088c92d91d93722c47f88163a8", false, true},
	} {
		t.Run(vector.id, func(t *testing.T) {
			message, err := hex.DecodeString(vector.message)
			if err != nil {
				t.Fatal(err)
			}
			signature, err := hex.DecodeString(vector.signature)
			if err != nil {
				t.Fatal(err)
			}
			if (VerifyRSAPSSSHA256(key, message, signature) == nil) != vector.pss {
				t.Fatal("independent PSS vector mismatch")
			}
			if (VerifyRSAPKCS1SHA256(key, message, signature) == nil) != vector.pkcs {
				t.Fatal("independent PKCS1 vector mismatch")
			}
		})
	}
}
