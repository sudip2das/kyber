package promise

import (
	"bytes"
	"testing"
	"reflect"

	"github.com/dedis/crypto/anon"
	"github.com/dedis/crypto/abstract"
	"github.com/dedis/crypto/config"
	"github.com/dedis/crypto/edwards"
	"github.com/dedis/crypto/nist"
	"github.com/dedis/crypto/random"
)

var keySuite   = nist.NewAES128SHA256P256()
var shareGroup = new(edwards.ExtendedCurve).Init(edwards.Param25519(), false)

var secretKey   = produceKeyPair()
var promiserKey = produceKeyPair()

var pt          = 10
var r           = 15
var numInsurers = 20

var insurerKeys = produceinsurerKeys()
var insurerList = produceinsurerList()

var basicPromise      = new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
var basicPromiseState = new(PromiseState).Init(*basicPromise)

func produceKeyPair() *config.KeyPair {
	keyPair := new(config.KeyPair)
	keyPair.Gen(keySuite, random.Stream)
	return keyPair
}

func produceinsurerKeys() []*config.KeyPair {
	newArray := make([]*config.KeyPair, numInsurers, numInsurers)
	for i := 0; i < numInsurers; i++ {
		newArray[i] = produceKeyPair()
	}
	return newArray
}

func produceinsurerList() []abstract.Point {
	newArray := make([]abstract.Point, numInsurers, numInsurers)
	for i := 0; i < numInsurers; i++ {
		newArray[i] = insurerKeys[i].Public
	}
	return newArray
}

// Tests that check whether a method panics can use this funcition
func deferTest(t *testing.T, message string) {
	if r := recover(); r == nil {
		t.Error(message)
	}
}

// Verifies that Init properly initalizes a new PromiseSignature object
func TestPromiseSignatureInit(t *testing.T) {
	sig := []byte("This is a test signature")
	p   := new(PromiseSignature).init(keySuite, sig)
	if p.suite != keySuite {
		t.Error("Suite not properly initialized.")
	}
	if !reflect.DeepEqual(sig, p.signature) {
		t.Error("Signature not properly initialized.")
	}
}

// Verifies that UnMarshalInit properly initalizes for unmarshalling
func TestPromiseSignatureUnMarshalInit(t *testing.T) {
	p := new(PromiseSignature).UnmarshalInit(keySuite)
	if p.suite != keySuite {
		t.Error("Suite not properly initialized.")
	}
}

// Verifies that PromiseSignature's marshalling code works
func TestPromiseSignatureBinaryMarshalling(t *testing.T) {
	// Tests BinaryMarshal, BinaryUnmarshal, and MarshalSize
	sig := basicPromise.Sign(numInsurers-1, insurerKeys[numInsurers-1])
	encodedSig, err := sig.MarshalBinary()
	if err != nil || len(encodedSig) != sig.MarshalSize() {
		t.Fatal("Marshalling failed: ", err,
			len(encodedSig) != sig.MarshalSize())
	}
	
	decodedSig := new(PromiseSignature).UnmarshalInit(keySuite)
	err         = decodedSig.UnmarshalBinary(encodedSig)
	if err != nil {
		t.Fatal("UnMarshalling failed: ", err)
	}
	if !sig.Equal(decodedSig) {
		t.Error("Decoded signature not equal to original")
	}
	if basicPromise.VerifySignature(numInsurers-1, decodedSig) != nil {
		t.Error("Decoded signature failed to be verified.")
	}
	
	// Tests MarshlTo and UnmarshalFrom
	sig2               := basicPromise.Sign(1, insurerKeys[1])
	bufWriter          := new(bytes.Buffer)
	bytesWritter, errs := sig2.MarshalTo(bufWriter)
	if bytesWritter != sig2.MarshalSize() || errs != nil {
		t.Fatal("MarshalTo failed: ", bytesWritter, err)
	}
	
	decodedSig2      := new(PromiseSignature).UnmarshalInit(keySuite)
	bufReader        := bytes.NewReader(bufWriter.Bytes())
	bytesRead, errs2 := decodedSig2.UnmarshalFrom(bufReader)
	if bytesRead != sig2.MarshalSize() || errs2 != nil {
		t.Fatal("UnmarshalFrom failed: ", bytesRead, errs2)
	}
	if sig2.MarshalSize() != decodedSig2.MarshalSize() {
		t.Error("MarshalSize of decoded and original differ: ",
			sig2.MarshalSize(), decodedSig2.MarshalSize())
	}
	if !sig2.Equal(decodedSig2) {
		t.Error("PromiseSignature read does not equal original")
	}
	if basicPromise.VerifySignature(1, decodedSig2) != nil {
		t.Error("Read signature failed to be verified.")
	}
	
}

// Verifies that Equal properly works for PromiseSignature objects
func TestPromiseSignatureEqual(t *testing.T) {
	sig := []byte("This is a test")
	p := new(PromiseSignature).init(keySuite, sig)
	if !p.Equal(p) {
		t.Error("PromiseSignature should equal itself.")
	}
	
	// Error cases
	p2 := new(PromiseSignature).init(nil, sig)	
	if p.Equal(p2) {
		t.Error("PromiseSignature's differ in suite.")
	}
	p2 = new(PromiseSignature).init(keySuite, nil)	
	if p.Equal(p2) {
		t.Error("PromiseSignature's differ in signature.")
	}
}

// Verifies that Init properly initalizes a new BlameProof object
func TestBlameProofInit(t *testing.T) {
	proof := []byte("This is a test")
	sig   := []byte("This too is a test")
	p     := new(PromiseSignature).init(keySuite, sig)
	bp    := new(BlameProof).init(keySuite, promiserKey.Public, proof, p)
	if keySuite != bp.suite  {
		t.Error("Suite not properly initialized.")
	}
	if !bp.diffieKey.Equal(promiserKey.Public) {
		t.Error("Diffie-Hellman key not properly initialized.")
	}
	if !reflect.DeepEqual(bp.diffieKeyProof, proof) {
		t.Error("Diffie-Hellman proof not properly initialized.")
	}
	if !p.Equal(&bp.signature) {
		t.Error("PromisSignature not properly initialized.")
	}
}

// Verifies that UnMarshalInit properly initalizes for unmarshalling
func TestBlameProofUnMarshalInit(t *testing.T) {
	bp := new(BlameProof).UnmarshalInit(keySuite)
	if bp.suite != keySuite {
		t.Error("BlameProof not properly initialized.")
	}
}

// Verifies that Equal properly works for PromiseSignature objects
func TestBlameProofEqual(t *testing.T) {
	p  := new(PromiseSignature).init(keySuite, []byte("Test"))
	bp := new(BlameProof).init(keySuite, promiserKey.Public, []byte("Test"), p)
	if !bp.Equal(bp) {
		t.Error("BlameProof should equal itself.")
	}
	
	// Error cases
	bp2 := new(BlameProof).init(nil, promiserKey.Public, []byte("Test"), p)
	if bp.Equal(bp2) {
		t.Error("BlameProof differ in key suites.")
	}
	bp2 = new(BlameProof).init(keySuite, keySuite.Point().Base(), []byte("Test"), p)
	if bp.Equal(bp2) {
		t.Error("BlameProof differ in diffie-keys.")
	}
	bp2 = new(BlameProof).init(keySuite, promiserKey.Public, []byte("Differ"), p)
	if bp.Equal(bp2) {
		t.Error("BlameProof differ in hash proof.")
	}
	p2 := new(PromiseSignature).init(keySuite, []byte("Differ"))
	bp2 = new(BlameProof).init(keySuite, promiserKey.Public, []byte("Test"), p2)
	if bp.Equal(bp2) {
		t.Error("BlameProof differ in signatures.")
	}
}

// Verifies that BlameProof's marshalling methods work properly.
func TestBlameProofBinaryMarshalling(t *testing.T) {
	// Create a bad promise object. That a blame proof would succeed.
	promise := new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	badKey  := insurerKeys[numInsurers-1]
	diffieBase := promise.shareSuite.Point().Mul(promiserKey.Public, badKey.Secret)
	badShare := promise.diffieHellmanEncrypt(badKey.Secret, diffieBase)
	promise.secrets[0] = badShare


	// Tests BinaryMarshal, BinaryUnmarshal, and MarshalSize
	bp,_ := promise.Blame(0, insurerKeys[0])
	encodedBp, err := bp.MarshalBinary()
	if err != nil || len(encodedBp) != bp.MarshalSize() {
		t.Fatal("Marshalling failed: ", err)
	}
	
	decodedBp := new(BlameProof).UnmarshalInit(keySuite)
	err        = decodedBp.UnmarshalBinary(encodedBp)
	if err != nil {
		t.Fatal("UnMarshalling failed: ", err)
	}
	if !bp.Equal(decodedBp) {
		t.Error("Decoded BlameProof not equal to original")
	}
	if bp.MarshalSize() != decodedBp.MarshalSize() {
		t.Error("MarshalSize of decoded and original differ: ",
			bp.MarshalSize(), decodedBp.MarshalSize())
	}
	if promise.VerifyBlame(0, decodedBp) != nil {
		t.Error("Decoded BlameProof failed to be verified.")
	}
	
	// Tests MarshlTo and UnmarshalFrom
	bp2, _ := basicPromise.Blame(0, insurerKeys[0])
	bufWriter := new(bytes.Buffer)
	bytesWritter, errs := bp2.MarshalTo(bufWriter)
	if bytesWritter != bp2.MarshalSize() || errs != nil {
		t.Fatal("MarshalTo failed: ", bytesWritter, err)
	}
	
	decodedBp2       := new(BlameProof).UnmarshalInit(keySuite)
	bufReader        := bytes.NewReader(bufWriter.Bytes())
	bytesRead, errs2 := decodedBp2.UnmarshalFrom(bufReader)
	if bytesRead != bp2.MarshalSize() || errs2 != nil {
		t.Fatal("UnmarshalFrom failed: ", bytesRead, errs2)
	}
	if bp2.MarshalSize() != decodedBp2.MarshalSize() {
		t.Error("MarshalSize of decoded and original differ: ",
			bp2.MarshalSize(), decodedBp2.MarshalSize())
	}
	if !bp2.Equal(decodedBp2) {
		t.Error("BlameProof read does not equal original")
	}
	if promise.VerifyBlame(0, decodedBp2) != nil {
		t.Error("Decoded BlameProof failed to be verified.")
	}
	
}

// Verifies that Init properly initalizes a new Promise object
func TestPromiseInit(t *testing.T) {

	// Verify that a promise can be initialized properly.
	promise := new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
		
	if promiserKey.Suite.String() != promise.shareSuite.String() ||
	   promise.t != pt || promise.r != r || promise.n != numInsurers ||
	   promise.pubKey != promiserKey.Public ||
	   len(promise.secrets)    != numInsurers {
		t.Error("Invalid initialization")	   
	}

	for i := 0 ; i < promise.n; i++ {
	
	   	if !insurerList[i].Equal(promise.insurers[i]) {
	   		t.Error("Public key for insurer not added:", i)
	   	}

		diffieBase := promise.shareSuite.Point().Mul(insurerList[i], promiserKey.Secret)
		share := promise.diffieHellmanDecrypt(promise.secrets[i], diffieBase)
		if !promise.pubPoly.Check(i, share) {
			t.Error("Polynomial Check failed for share ", i)
		}
	}
	
	// Error handling
	
	// Check that Init panics if n < t
	test := func() {
		defer deferTest(t, "Init should have panicked.")
		new(Promise).ConstructPromise(secretKey, promiserKey, pt, r,
			[]abstract.Point{promiserKey.Public})
	}

	test()
	
	
	// Check that r is reset properly when r < t.
	promise = new(Promise).ConstructPromise(secretKey, promiserKey, pt, pt-20, insurerList)
	if promise.r < pt || promise.r > numInsurers {
		t.Error("Invalid r allowed for r < t.")
	}


	// Check that r is reset properly when r > n.
	promise = new(Promise).ConstructPromise(secretKey, promiserKey, pt, numInsurers+20, insurerList)
	if  promise.r < pt || promise.r > numInsurers {
		t.Error("Invalid r allowed for r > n.")
	}
}

// Verifies that UnMarshalInit properly initalizes for unmarshalling
func TestPromiseUnMarshalInit(t *testing.T) {
	p := new(Promise).UnmarshalInit(keySuite)
	if p.shareSuite != keySuite {
		t.Error("Promise not properly initialized.")
	}
}

// Tests that PromiseVerify properly rules out invalidly constructed Promise's
func TestPromiseVerifyPromise(t *testing.T) {
	promise  := new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	if promise.VerifyPromise(promiserKey.Public) != nil {
		t.Error("Promise is valid")
	}

	promise   = new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	promise.t = promise.n +1
	if promise.VerifyPromise(promiserKey.Public) == nil {
		t.Error("Promise is invalid: t > n")
	}

	promise   = new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	promise.t = promise.r +1
	if promise.VerifyPromise(promiserKey.Public) == nil {
		t.Error("Promise is invalid: t > r")
	}
	
	promise   = new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	promise.r = promise.n +1
	if promise.VerifyPromise(promiserKey.Public) == nil {
		t.Error("Promise is invalid: n > r")
	}
	
	promise   = new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	promise.pubKey = insurerList[0]
	if promise.VerifyPromise(promiserKey.Public) == nil {
		t.Error("Promise is invalid: the public key is wrong")
	}
	
	promise   = new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	promise.insurers = []abstract.Point{}
	if promise.VerifyPromise(promiserKey.Public) == nil {
		t.Error("Promise is invalid: insurers list is the wrong length")
	}
	
	promise   = new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	promise.secrets = []abstract.Secret{}
	if promise.VerifyPromise(promiserKey.Public) == nil {
		t.Error("Promise is invalid: secrets list is the wrong length")
	}
}


// Tests that encrypting a secret with a diffie-hellman shared key and then
// decrypting it succeeds.
func TestPromiseDiffieHellmanEncryptDecrypt(t *testing.T) {
	// key2 and promiserKey will be the two parties. The secret they are
	// share is the private key of secretKey
	key2      := produceKeyPair()
	secretKey := produceKeyPair()
	
	diffieBaseBasic := basicPromise.shareSuite.Point().Mul(key2.Public, promiserKey.Secret)
	encryptedSecret := basicPromise.diffieHellmanEncrypt(secretKey.Secret, diffieBaseBasic)


	diffieBaseKey2 := basicPromise.shareSuite.Point().Mul(promiserKey.Public, key2.Secret)
	secret := basicPromise.diffieHellmanDecrypt(encryptedSecret, diffieBaseKey2)

	if !secret.Equal(secretKey.Secret) {
		t.Error("Diffie-Hellman encryption/decryption failed.")
	}
}

// Tests that insurers can properly verify their share. Make sure that
// verification fails if the proper credentials are not supplied (aka Diffie-
// Hellman decryption failed).
func TestPromiseVerifyShare(t *testing.T) {
	if basicPromise.VerifyShare(0, insurerKeys[0]) != nil{
		t.Error("The share should have been verified")
	}
	
	// Make sure the wrong index and key pair fail.
	if basicPromise.VerifyShare(-1, insurerKeys[0]) == nil{
		t.Error("The share should not have been valid. Index is negative")
	}

	// Make sure the wrong index and key pair fail.
	if basicPromise.VerifyShare(basicPromise.n, insurerKeys[0]) == nil {
		t.Error("The share should not have been valid. Index >= n")
	}
	
	// Make sure the wrong index and key pair fail.
	if basicPromise.VerifyShare(numInsurers-1, insurerKeys[0]) == nil {
		t.Error("The share should not have been valid.")
	}
}

// Verify that the promise can produce a valid signature and then verify it.
// In short, all signatures produced by the sign method should be accepted.
func TestPromiseSignAndVerify(t *testing.T) {
	for i := 0 ; i < numInsurers; i++ {
		sig := basicPromise.Sign(i, insurerKeys[i])
		if basicPromise.VerifySignature(i, sig) != nil {
			t.Error("Signature failed to be validated")
		}
	}
}

// Produces a bad signature that has a malformed approve message
func produceSigWithBadMessage() *PromiseSignature {
	set        := anon.Set{insurerKeys[0].Public}
	approveMsg := "Bad message"
	digSig     := anon.Sign(insurerKeys[0].Suite, random.Stream, []byte(approveMsg),
		     set, nil, 0, insurerKeys[0].Secret)
		     
	return new(PromiseSignature).init(insurerKeys[0].Suite, digSig)
}


// Verify that mallformed signatures are not accepted.
func TestPromiseVerifySignature(t *testing.T) {
	// Fail if the signature is not the specially formatted approve message.
	if basicPromise.VerifySignature(0, produceSigWithBadMessage()) == nil {
		t.Error("Signature has a bad message and should be rejected.")
	}
	
	// Fail if a valid signature is applied to the wrong share.
	sig := basicPromise.Sign(0, insurerKeys[0])
	if basicPromise.VerifySignature(numInsurers-1, sig) == nil {
		t.Error("Signature is for the wrong share.")
	}

	// Fail if index is negative
	if basicPromise.VerifySignature(-1, sig)  == nil{
		t.Error("Error: Index < 0")
	}

	// Fail if index >= n
	if basicPromise.VerifySignature(basicPromise.n, sig)  == nil{
		t.Error("Error: Index >= n")
	}
	
	// Should return false if passed nil
	sig.signature = nil
	if basicPromise.VerifySignature(0, sig) == nil {
		t.Error("Error: Signature is nil")
	}
}

// Verify that insurer secret shares can be revealed properly and verified.
func TestPromiseRevealShareAndShareVerify(t *testing.T) {

	promiseShare := basicPromise.RevealShare(0, insurerKeys[0])
	if basicPromise.VerifyRevealedShare(0, promiseShare) != nil {
		t.Error("The share should have been marked as valid")
	}
	
	// Error Handling
	badShare := basicPromise.RevealShare(0, insurerKeys[0])
	if basicPromise.VerifyRevealedShare(-10, badShare) == nil {
		t.Error("The index provided is too low.")
	}


	badShare = basicPromise.RevealShare(0, insurerKeys[0])
	if basicPromise.VerifyRevealedShare(numInsurers + 20, badShare) == nil {
		t.Error("The index provided is too high.")
	}
	
	badShare = basicPromise.RevealShare(0, insurerKeys[0])
	if basicPromise.VerifyRevealedShare(0, insurerKeys[0].Secret) == nil {
		t.Error("The share provided is bad.")
	}
}

// Verify that insurers can properly create and verify blame proofs
func TestPromiseBlameAndVerify(t *testing.T) {

	// Create a bad promise object. Create a new secret that will fail the
	// the public polynomial check. 
	promise := new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	badKey := insurerKeys[numInsurers-1]
	
	diffieBase := promise.shareSuite.Point().Mul(promiserKey.Public, badKey.Secret)
	badShare := promise.diffieHellmanEncrypt(badKey.Secret, diffieBase)
	
	promise.secrets[0] = badShare


	validProof, err := promise.Blame(0, insurerKeys[0])
	if err != nil {
		t.Fatal("Blame failed to be properly constructed")
	}

	if promise.VerifyBlame(0, validProof) != nil {
		t.Error("The proof is valid and should be accepted.")
	}

	// Error handling
	goodPromiseShare, _ := basicPromise.Blame(0, insurerKeys[0])
	if basicPromise.VerifyBlame(0, goodPromiseShare) == nil {
		t.Error("Invalid blame: the share is actually good.")
	}

	if basicPromise.VerifyBlame(-10, goodPromiseShare) == nil {
		t.Error("The i index is below 0")
	}

	if basicPromise.VerifyBlame(numInsurers+20, goodPromiseShare) == nil {
		t.Error("The i index is below above n")
	}

	badProof, _ := basicPromise.Blame(0, insurerKeys[0])
	badProof.diffieKeyProof = []byte("This is an invalid zero-knowledge proof")
	if basicPromise.VerifyBlame(0, badProof) == nil {
		t.Error("Invalid blame. The verification of the diffie-key proof is bad.")
	}

	badSignature, _ := basicPromise.Blame(0, insurerKeys[0])
	badSignature.signature = *promise.Sign(1, insurerKeys[1])
	if basicPromise.VerifyBlame(0, badSignature)  == nil {
		t.Error("Invalid blame. The signature is bad.")
	}
}

// Verifies that Equal properly works for PromiseSignature objects
func TestPromiseEqual(t *testing.T) {
	// Make sure promise equals basicPromise to make testing error cases
	// below valid (if promise never == basicPromise, the error cases are
	// trivially true). Secrets and the public polynomial must be set
	// equal in each case to make sure that promise and basicPromise are
	// equal.
	promise := new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	promise.secrets = basicPromise.secrets
	promise.pubPoly = basicPromise.pubPoly
	if !basicPromise.Equal(promise) {
		t.Error("Promises should be equal.")
	}

	
	// Error cases
	promise = new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	promise.secrets = basicPromise.secrets
	promise.pubPoly = basicPromise.pubPoly
	promise.shareSuite = nil
	if basicPromise.Equal(promise) {
		t.Error("The shareSuite's are not equal")
	}

	promise = new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	promise.secrets = basicPromise.secrets
	promise.pubPoly = basicPromise.pubPoly
	promise.n = 0
	if basicPromise.Equal(promise) {
		t.Error("The n's are not equal")
	}

	promise = new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	promise.secrets = basicPromise.secrets
	promise.pubPoly = basicPromise.pubPoly
	promise.t = 0
	if basicPromise.Equal(promise) {
		t.Error("The t's are not equal")
	}

	promise = new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	promise.secrets = basicPromise.secrets
	promise.pubPoly = basicPromise.pubPoly
	promise.pubKey = keySuite.Point().Base()
	if basicPromise.Equal(promise) {
		t.Error("The public keys are not equal")
	}

	promise = new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	promise.secrets = basicPromise.secrets
	if basicPromise.Equal(promise) {
		t.Error("The public polynomials are not equal")
	}


	promise = new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	promise.secrets = basicPromise.secrets
	promise.pubPoly = basicPromise.pubPoly
	promise.insurers = make([]abstract.Point, promise.n, promise.n)
	copy(promise.insurers, insurerList)
	promise.insurers[numInsurers-1] = keySuite.Point().Base()
	if basicPromise.Equal(promise) {
		t.Error("The insurers array are not equal")
	}

	promise = new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	promise.pubPoly = basicPromise.pubPoly
	if basicPromise.Equal(promise) {
		t.Error("The secrets array are not equal")
	}
}



// Verifies that UnMarshalInit properly initalizes for unmarshalling
func TestPromiseBinaryMarshalling(t *testing.T) {

	// Tests BinaryMarshal, BinaryUnmarshal, and MarshalSize
	encodedP, err := basicPromise.MarshalBinary()
	if err != nil || len(encodedP) != basicPromise.MarshalSize() {
		t.Fatal("Marshalling failed: ", err)
	}
	
	decodedP := new(Promise).UnmarshalInit(keySuite)
	err = decodedP.UnmarshalBinary(encodedP)
	if err != nil {
		t.Fatal("UnMarshalling failed: ", err)
	}
	if !basicPromise.Equal(decodedP) {
		t.Error("Decoded BlameProof not equal to original")
	}
	
	// Tests MarshlTo and UnmarshalFrom
	bufWriter := new(bytes.Buffer)
	bytesWritter, errs := basicPromise.MarshalTo(bufWriter)
	
	if bytesWritter != basicPromise.MarshalSize() || errs != nil {
		t.Fatal("MarshalTo failed: ", bytesWritter, err)
	}
	
	decodedP2 := new(Promise).UnmarshalInit(keySuite)
	bufReader := bytes.NewReader(bufWriter.Bytes())
	bytesRead, errs2 := decodedP2.UnmarshalFrom(bufReader)
	if bytesRead != decodedP2.MarshalSize() ||
	   basicPromise.MarshalSize() != decodedP2.MarshalSize() ||
	   errs2 != nil {
		t.Fatal("UnmarshalFrom failed: ", bytesRead, errs2)
	}
	if !basicPromise.Equal(decodedP2) {
		t.Error("BlameProof read does not equal original")
	}
}


// Verifies that Init properly initalizes a new PromiseState object
func TestPromiseStateInit(t *testing.T) {

	promiseState := new(PromiseState).Init(*basicPromise)
	
	if //!basicPromise.Equal(promiseState.Promise) || <-- Once I write Equal
	   len(promiseState.signatures) != numInsurers {
		t.Error("Invalid initialization")	   
	}
}

// Verify that Promise and PromiseState can produce a valid signature and then verify it.
func TestPromiseStateAddSignature(t *testing.T) {

	promise := new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	promiseState := new(PromiseState).Init(*promise)

	// Verify that all validly produced signatures can be added.
	for i := 0 ; i < numInsurers; i++ {
		sig := promise.Sign(i, insurerKeys[i])
		promiseState.AddSignature(i, sig)
		
		if !sig.Equal(promiseState.signatures[i]) {
			t.Error("Signature failed to be added")
		}
	}
}

// Verify that PromiseState can add blames.
func TestPromiseStateAddBlame(t *testing.T) {

	promise := new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	promiseState := new(PromiseState).Init(*promise)

	// Ensure that blames can be added.
	for i := 0 ; i < numInsurers; i++ {
		bproof, _ := promise.Blame(i, insurerKeys[i])
		promiseState.AddBlameProof(i, bproof)
		
		if !bproof.Equal(promiseState.blames[i]) {
			t.Error("Blame failed to be added")
		}
	}
}


// Verify that once r signatures have been added, the promise becomes valid.
func TestPromiseStatePromiseCertified(t *testing.T) {

	promise := new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	promiseState := new(PromiseState).Init(*promise)

	for i := 0 ; i < numInsurers; i++ {
		promiseState.AddSignature(i, promise.Sign(i, insurerKeys[i]))
	
		// Insure that invalidly added proofs do not distort the proof.
		bproof, _ := promise.Blame(i, insurerKeys[i])
		promiseState.AddBlameProof(i, bproof)
		
		err := promiseState.PromiseCertified(promiserKey.Public)
		if i < r-1 && err == nil {
			t.Error("Not enough signtures have been added yet", i, r)
		} else if i >= r-1 && err != nil {
			t.Error("Promise should be valid now.")
			t.Error(promiseState.PromiseCertified(promiserKey.Public))
		}
	}

	promise      = new(Promise).ConstructPromise(secretKey, promiserKey, pt, r, insurerList)
	promiseState = new(PromiseState).Init(*promise)
	
	promise.secrets[0] = promise.shareSuite.Secret()

	for i := 0 ; i < numInsurers; i++ {
		promiseState.AddSignature(i, promise.Sign(i, insurerKeys[i]))
	
		// Insure that invalidly added proofs do not distort the proof.
		bproof, _ := promise.Blame(i, insurerKeys[i])
		promiseState.AddBlameProof(i, bproof)

		if promiseState.PromiseCertified(promiserKey.Public) == nil {
			t.Error("Not enough signtures have been added yet", i, r)
		}
	}
}

// Tests all the string functions. Simply calls them to make sure they return.
func TestString(t *testing.T) {
	sig := basicPromise.Sign(0, insurerKeys[0])
	sig.String()
	
	bp, _ := basicPromise.Blame(0, insurerKeys[0])
	bp.String()

	basicPromise.String()
}
