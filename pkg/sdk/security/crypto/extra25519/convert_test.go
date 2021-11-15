// Copyright 2021 Thibault NORMAND
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package extra25519_test

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/elastic/harp/pkg/sdk/security/crypto/extra25519"
)

func ExamplePrivateKeyToCurve25519() {
	dumper := hex.Dumper(os.Stdout)

	// really silly seed for reproduciability
	seed := make([]byte, 32)

	fmt.Println("seed:")
	dumper.Write(seed)

	_, private, err := ed25519.GenerateKey(bytes.NewReader(seed))
	fatal(err)

	fmt.Println("private ed25519:")
	dumper.Write(private)

	var curvPriv [32]byte
	extra25519.PrivateKeyToCurve25519(&curvPriv, private)

	fmt.Println("private curve25519:")
	dumper.Write(curvPriv[:])

	// Output:
	// seed:
	// 00000000  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	// 00000010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	// private ed25519:
	// 00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	// 00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	// 00000040  3b 6a 27 bc ce b6 a4 2d  62 a3 a8 d0 2a 6f 0d 73  |;j'....-b...*o.s|
	// 00000050  65 32 15 77 1d e2 43 a6  3a c0 48 a1 8b 59 da 29  |e2.w..C.:.H..Y.)|
	// private curve25519:
	// 00000060  50 46 ad c1 db a8 38 86  7b 2b bb fd d0 c3 42 3e  |PF....8.{+....B>|
	// 00000070  58 b5 79 70 b5 26 7a 90  f5 79 60 92 4a 87 f1 56  |X.yp.&z..y`.J..V|
}

func ExamplePublicKeyToCurve25519() {
	dumper := hex.Dumper(os.Stdout)

	// really silly seed for reproduciability
	seed := make([]byte, 32)

	fmt.Println("seed:")
	dumper.Write(seed)

	public, _, err := ed25519.GenerateKey(bytes.NewReader(seed))
	fatal(err)

	fmt.Println("public ed25519:")
	dumper.Write(public)

	var curvPub [32]byte
	extra25519.PublicKeyToCurve25519(&curvPub, public)

	fmt.Println("public curve25519:")
	dumper.Write(curvPub[:])

	// Output:
	// seed:
	// 00000000  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	// 00000010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	// public ed25519:
	// 00000020  3b 6a 27 bc ce b6 a4 2d  62 a3 a8 d0 2a 6f 0d 73  |;j'....-b...*o.s|
	// 00000030  65 32 15 77 1d e2 43 a6  3a c0 48 a1 8b 59 da 29  |e2.w..C.:.H..Y.)|
	// public curve25519:
	// 00000040  5b f5 5c 73 b8 2e be 22  be 80 f3 43 06 67 af 57  |[.\s..."...C.g.W|
	// 00000050  0f ae 25 56 a6 41 5e 6b  30 d4 06 53 00 aa 94 7d  |..%V.A^k0..S...}|
}

func fatal(err error) {
	if err != nil {
		panic(err)
	}
}
