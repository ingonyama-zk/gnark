// Copyright 2020 ConsenSys Software Inc.
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

// Code generated by gnark DO NOT EDIT

package plonk

import (
	curve "github.com/consensys/gnark-crypto/ecc/bls12-381"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"

	"errors"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr/iop"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr/kzg"
	"io"
)

// WriteRawTo writes binary encoding of Proof to w without point compression
func (proof *Proof) WriteRawTo(w io.Writer) (int64, error) {
	return proof.writeTo(w, curve.RawEncoding())
}

// WriteTo writes binary encoding of Proof to w with point compression
func (proof *Proof) WriteTo(w io.Writer) (int64, error) {
	return proof.writeTo(w)
}

func (proof *Proof) writeTo(w io.Writer, options ...func(*curve.Encoder)) (int64, error) {
	enc := curve.NewEncoder(w, options...)

	toEncode := []interface{}{
		&proof.LRO[0],
		&proof.LRO[1],
		&proof.LRO[2],
		&proof.Z,
		&proof.H[0],
		&proof.H[1],
		&proof.H[2],
		&proof.BatchedProof.H,
		proof.BatchedProof.ClaimedValues,
		&proof.ZShiftedOpening.H,
		&proof.ZShiftedOpening.ClaimedValue,
		proof.Bsb22Commitments,
	}

	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			return enc.BytesWritten(), err
		}
	}

	return enc.BytesWritten(), nil
}

// ReadFrom reads binary representation of Proof from r
func (proof *Proof) ReadFrom(r io.Reader) (int64, error) {
	dec := curve.NewDecoder(r)
	toDecode := []interface{}{
		&proof.LRO[0],
		&proof.LRO[1],
		&proof.LRO[2],
		&proof.Z,
		&proof.H[0],
		&proof.H[1],
		&proof.H[2],
		&proof.BatchedProof.H,
		&proof.BatchedProof.ClaimedValues,
		&proof.ZShiftedOpening.H,
		&proof.ZShiftedOpening.ClaimedValue,
		&proof.Bsb22Commitments,
	}

	for _, v := range toDecode {
		if err := dec.Decode(v); err != nil {
			return dec.BytesRead(), err
		}
	}

	if proof.Bsb22Commitments == nil {
		proof.Bsb22Commitments = []kzg.Digest{}
	}

	return dec.BytesRead(), nil
}

// WriteTo writes binary encoding of ProvingKey to w
func (pk *ProvingKey) WriteTo(w io.Writer) (n int64, err error) {
	return pk.writeTo(w, true)
}

// WriteRawTo writes binary encoding of ProvingKey to w without point compression
func (pk *ProvingKey) WriteRawTo(w io.Writer) (n int64, err error) {
	return pk.writeTo(w, false)
}

func (pk *ProvingKey) writeTo(w io.Writer, withCompression bool) (n int64, err error) {
	// encode the verifying key
	if withCompression {
		n, err = pk.Vk.WriteTo(w)
	} else {
		n, err = pk.Vk.WriteRawTo(w)
	}
	if err != nil {
		return
	}

	// fft domains
	n2, err := pk.Domain[0].WriteTo(w)
	if err != nil {
		return
	}
	n += n2

	n2, err = pk.Domain[1].WriteTo(w)
	if err != nil {
		return
	}
	n += n2

	// KZG key
	if withCompression {
		n2, err = pk.Kzg.WriteTo(w)
	} else {
		n2, err = pk.Kzg.WriteRawTo(w)
	}
	if err != nil {
		return
	}
	n += n2

	// sanity check len(Permutation) == 3*int(pk.Domain[0].Cardinality)
	if len(pk.trace.S) != (3 * int(pk.Domain[0].Cardinality)) {
		return n, errors.New("invalid permutation size, expected 3*domain cardinality")
	}

	enc := curve.NewEncoder(w)
	// note: type Polynomial, which is handled by default binary.Write(...) op and doesn't
	// encode the size (nor does it convert from Montgomery to Regular form)
	// so we explicitly transmit []fr.Element
	toEncode := []interface{}{
		pk.trace.Ql.Coefficients(),
		pk.trace.Qr.Coefficients(),
		pk.trace.Qm.Coefficients(),
		pk.trace.Qo.Coefficients(),
		pk.trace.Qk.Coefficients(),
		coefficients(pk.trace.Qcp),
		pk.trace.S1.Coefficients(),
		pk.trace.S2.Coefficients(),
		pk.trace.S3.Coefficients(),
		pk.trace.S,
	}

	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			return n + enc.BytesWritten(), err
		}
	}

	return n + enc.BytesWritten(), nil
}

// ReadFrom reads from binary representation in r into ProvingKey
func (pk *ProvingKey) ReadFrom(r io.Reader) (int64, error) {
	return pk.readFrom(r, true)
}

// UnsafeReadFrom reads from binary representation in r into ProvingKey without subgroup checks
func (pk *ProvingKey) UnsafeReadFrom(r io.Reader) (int64, error) {
	return pk.readFrom(r, false)
}

func (pk *ProvingKey) readFrom(r io.Reader, withSubgroupChecks bool) (int64, error) {
	pk.Vk = &VerifyingKey{}
	n, err := pk.Vk.ReadFrom(r)
	if err != nil {
		return n, err
	}

	n2, err, chDomain0 := pk.Domain[0].AsyncReadFrom(r)
	n += n2
	if err != nil {
		return n, err
	}

	n2, err, chDomain1 := pk.Domain[1].AsyncReadFrom(r)
	n += n2
	if err != nil {
		return n, err
	}

	if withSubgroupChecks {
		n2, err = pk.Kzg.ReadFrom(r)
	} else {
		n2, err = pk.Kzg.UnsafeReadFrom(r)
	}
	n += n2
	if err != nil {
		return n, err
	}

	pk.trace.S = make([]int64, 3*pk.Domain[0].Cardinality)

	dec := curve.NewDecoder(r)

	var ql, qr, qm, qo, qk, s1, s2, s3 []fr.Element
	var qcp [][]fr.Element

	// TODO @gbotrel: this is a bit ugly, we should probably refactor this.
	// The order of the variables is important, as it matches the order in which they are
	// encoded in the WriteTo(...) method.

	// Note: instead of calling dec.Decode(...) for each of the above variables,
	// we call AsyncReadFrom when possible which allows to consume bytes from the reader
	// and perform the decoding in parallel

	type v struct {
		data  *fr.Vector
		chErr chan error
	}

	vectors := make([]v, 8)
	vectors[0] = v{data: (*fr.Vector)(&ql)}
	vectors[1] = v{data: (*fr.Vector)(&qr)}
	vectors[2] = v{data: (*fr.Vector)(&qm)}
	vectors[3] = v{data: (*fr.Vector)(&qo)}
	vectors[4] = v{data: (*fr.Vector)(&qk)}
	vectors[5] = v{data: (*fr.Vector)(&s1)}
	vectors[6] = v{data: (*fr.Vector)(&s2)}
	vectors[7] = v{data: (*fr.Vector)(&s3)}

	// read ql, qr, qm, qo, qk
	for i := 0; i < 5; i++ {
		n2, err, ch := vectors[i].data.AsyncReadFrom(r)
		n += n2
		if err != nil {
			return n, err
		}
		vectors[i].chErr = ch
	}

	// read qcp
	if err := dec.Decode(&qcp); err != nil {
		return n + dec.BytesRead(), err
	}

	// read lqk, s1, s2, s3
	for i := 5; i < 8; i++ {
		n2, err, ch := vectors[i].data.AsyncReadFrom(r)
		n += n2
		if err != nil {
			return n, err
		}
		vectors[i].chErr = ch
	}

	// read pk.Trace.S
	if err := dec.Decode(&pk.trace.S); err != nil {
		return n + dec.BytesRead(), err
	}

	// wait for all AsyncReadFrom(...) to complete
	for i := range vectors {
		if err := <-vectors[i].chErr; err != nil {
			return n, err
		}
	}

	canReg := iop.Form{Basis: iop.Canonical, Layout: iop.Regular}
	pk.trace.Ql = iop.NewPolynomial(&ql, canReg)
	pk.trace.Qr = iop.NewPolynomial(&qr, canReg)
	pk.trace.Qm = iop.NewPolynomial(&qm, canReg)
	pk.trace.Qo = iop.NewPolynomial(&qo, canReg)
	pk.trace.Qk = iop.NewPolynomial(&qk, canReg)
	pk.trace.S1 = iop.NewPolynomial(&s1, canReg)
	pk.trace.S2 = iop.NewPolynomial(&s2, canReg)
	pk.trace.S3 = iop.NewPolynomial(&s3, canReg)

	pk.trace.Qcp = make([]*iop.Polynomial, len(qcp))
	for i := range qcp {
		pk.trace.Qcp[i] = iop.NewPolynomial(&qcp[i], canReg)
	}

	// wait for FFT to be precomputed
	<-chDomain0
	<-chDomain1

	pk.computeLagrangeCosetPolys()

	return n + dec.BytesRead(), nil

}

// WriteTo writes binary encoding of VerifyingKey to w
func (vk *VerifyingKey) WriteTo(w io.Writer) (n int64, err error) {
	return vk.writeTo(w)
}

// WriteRawTo writes binary encoding of VerifyingKey to w without point compression
func (vk *VerifyingKey) WriteRawTo(w io.Writer) (int64, error) {
	return vk.writeTo(w, curve.RawEncoding())
}

func (vk *VerifyingKey) writeTo(w io.Writer, options ...func(*curve.Encoder)) (n int64, err error) {
	enc := curve.NewEncoder(w)

	toEncode := []interface{}{
		vk.Size,
		&vk.SizeInv,
		&vk.Generator,
		vk.NbPublicVariables,
		&vk.CosetShift,
		&vk.S[0],
		&vk.S[1],
		&vk.S[2],
		&vk.Ql,
		&vk.Qr,
		&vk.Qm,
		&vk.Qo,
		&vk.Qk,
		vk.Qcp,
		&vk.Kzg.G1,
		&vk.Kzg.G2[0],
		&vk.Kzg.G2[1],
		vk.CommitmentConstraintIndexes,
	}

	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			return enc.BytesWritten(), err
		}
	}

	return enc.BytesWritten(), nil
}

// UnsafeReadFrom reads from binary representation in r into VerifyingKey.
// Current implementation is a passthrough to ReadFrom
func (vk *VerifyingKey) UnsafeReadFrom(r io.Reader) (int64, error) {
	return vk.ReadFrom(r)
}

// ReadFrom reads from binary representation in r into VerifyingKey
func (vk *VerifyingKey) ReadFrom(r io.Reader) (int64, error) {
	dec := curve.NewDecoder(r)
	toDecode := []interface{}{
		&vk.Size,
		&vk.SizeInv,
		&vk.Generator,
		&vk.NbPublicVariables,
		&vk.CosetShift,
		&vk.S[0],
		&vk.S[1],
		&vk.S[2],
		&vk.Ql,
		&vk.Qr,
		&vk.Qm,
		&vk.Qo,
		&vk.Qk,
		&vk.Qcp,
		&vk.Kzg.G1,
		&vk.Kzg.G2[0],
		&vk.Kzg.G2[1],
		&vk.CommitmentConstraintIndexes,
	}

	for _, v := range toDecode {
		if err := dec.Decode(v); err != nil {
			return dec.BytesRead(), err
		}
	}

	if vk.Qcp == nil {
		vk.Qcp = []kzg.Digest{}
	}

	return dec.BytesRead(), nil
}
