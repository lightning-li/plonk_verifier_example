package main

import (
	"bytes"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/test"
	"math/big"
	"os"
)

type DemoCircuit struct {
	X frontend.Variable
	Y frontend.Variable `gnark:",public"`
}

func (c *DemoCircuit) Define(api frontend.API) error {
	tmp := api.Mul(c.X, c.X)
	api.AssertIsEqual(tmp, c.Y)
	return nil
}

type PlonkFormattedProof struct {
	WireCommitments               [3][2]*big.Int
	GrandProductCommitment        [2]*big.Int
	QuotientPolyCommitments       [3][2]*big.Int
	WireValuesAtZeta              [3]*big.Int
	GrandProductAtZetaOmega       *big.Int
	QuotientPolynomialAtZeta      *big.Int
	LinearizationPolynomialAtZeta *big.Int
	PermutationPolynomialsAtZeta  [2]*big.Int
	OpeningAtZetaProof            [2]*big.Int
	OpeningAtZetaOmegaProof       [2]*big.Int
}

func (p *PlonkFormattedProof) ConvertToArray(res *[]*big.Int) {
	for i := 0; i < 3; i++ {
		*res = append(*res, p.WireCommitments[i][:]...)
	}
	*res = append(*res, p.GrandProductCommitment[:]...)
	for i := 0; i < 3; i++ {
		*res = append(*res, p.QuotientPolyCommitments[i][:]...)
	}
	*res = append(*res, p.WireValuesAtZeta[:]...)
	*res = append(*res, p.GrandProductAtZetaOmega)
	*res = append(*res, p.QuotientPolynomialAtZeta)
	*res = append(*res, p.LinearizationPolynomialAtZeta)
	*res = append(*res, p.PermutationPolynomialsAtZeta[:]...)
	*res = append(*res, p.OpeningAtZetaProof[:]...)
	*res = append(*res, p.OpeningAtZetaOmegaProof[:]...)
}

func FormatPlonkProof(oProof plonk.Proof) (proof *PlonkFormattedProof, err error) {
	proof = new(PlonkFormattedProof)
	const fpSize = 32
	var buf bytes.Buffer
	_, err = oProof.WriteTo(&buf)

	if err != nil {
		return nil, err
	}
	proofBytes := buf.Bytes()
	index := 0
	var g1point bn254.G1Affine
	for i := 0; i < 3; i++ {
		g1point.SetBytes(proofBytes[fpSize*index : fpSize*(index+1)])
		uncompressed := g1point.RawBytes()
		for j := 0; j < 2; j++ {
			proof.WireCommitments[i][j] = new(big.Int).SetBytes(uncompressed[fpSize*j : fpSize*(j+1)])
		}
		index += 1
	}

	g1point.SetBytes(proofBytes[fpSize*index : fpSize*(index+1)])
	uncompressed := g1point.RawBytes()
	proof.GrandProductCommitment[0] = new(big.Int).SetBytes(uncompressed[0:fpSize])
	proof.GrandProductCommitment[1] = new(big.Int).SetBytes(uncompressed[fpSize : fpSize*2])
	index += 1

	for i := 0; i < 3; i++ {
		g1point.SetBytes(proofBytes[fpSize*index : fpSize*(index+1)])
		uncompressed := g1point.RawBytes()
		for j := 0; j < 2; j++ {
			proof.QuotientPolyCommitments[i][j] = new(big.Int).SetBytes(uncompressed[fpSize*j : fpSize*(j+1)])
		}
		index += 1
	}

	g1point.SetBytes(proofBytes[fpSize*index : fpSize*(index+1)])
	uncompressed = g1point.RawBytes()
	proof.OpeningAtZetaProof[0] = new(big.Int).SetBytes(uncompressed[0:fpSize])
	proof.OpeningAtZetaProof[1] = new(big.Int).SetBytes(uncompressed[fpSize : fpSize*2])
	index += 1

	// plonk proof write len(ClaimedValues) which is 4 bytes
	offset := 4
	proof.QuotientPolynomialAtZeta = new(big.Int).SetBytes(proofBytes[offset+fpSize*index : offset+fpSize*(index+1)])
	index += 1

	proof.LinearizationPolynomialAtZeta = new(big.Int).SetBytes(proofBytes[offset+fpSize*index : offset+fpSize*(index+1)])
	index += 1

	for i := 0; i < 3; i++ {
		proof.WireValuesAtZeta[i] = new(big.Int).SetBytes(proofBytes[offset+fpSize*index : offset+fpSize*(index+1)])
		index += 1
	}

	for i := 0; i < 2; i++ {
		proof.PermutationPolynomialsAtZeta[i] = new(big.Int).SetBytes(proofBytes[offset+fpSize*index : offset+fpSize*(index+1)])
		index += 1
	}

	g1point.SetBytes(proofBytes[offset+fpSize*index : offset+fpSize*(index+1)])
	uncompressed = g1point.RawBytes()
	proof.OpeningAtZetaOmegaProof[0] = new(big.Int).SetBytes(uncompressed[0:fpSize])
	proof.OpeningAtZetaOmegaProof[1] = new(big.Int).SetBytes(uncompressed[fpSize : fpSize*2])
	index += 1

	proof.GrandProductAtZetaOmega = new(big.Int).SetBytes(proofBytes[offset+fpSize*index : offset+fpSize*(index+1)])
	return proof, nil
}

func main() {
	var c DemoCircuit
	oScs, err := frontend.Compile(ecc.BN254, scs.NewBuilder, &c, frontend.IgnoreUnconstrainedInputs())
	if err != nil {
		fmt.Println("err is ", err.Error())
		return
	}
	srs, _ := test.NewKZGSRS(oScs)
	pk, vk, err := plonk.Setup(oScs, srs)
	f, err := os.Create("PlonkVerifier" + ".sol")
	if err != nil {
		panic(err)
	}
	vk.ExportSolidity(f)

	var wit DemoCircuit
	wit.X = 11
	wit.Y = 121
	witness, _ := frontend.NewWitness(&wit, ecc.BN254)
	proof, err := plonk.Prove(oScs, pk, witness)
	if err != nil {
		panic(err)
	}

	var verifyWit DemoCircuit
	verifyWit.Y = 121
	verifyWitness, _ := frontend.NewWitness(&verifyWit, ecc.BN254, frontend.PublicOnly())
	err = plonk.Verify(proof, vk, verifyWitness)
	if err != nil {
		fmt.Println("verify proof failed. err is ", err.Error())
		return
	} else {
		fmt.Println("verify proof successfully")
	}
	fmt.Println("================================================")
	fmt.Printf("public inputs: [%d]\n", verifyWit.Y)
	formatProof, _ := FormatPlonkProof(proof)
	var res []*big.Int
	formatProof.ConvertToArray(&res)
	fmt.Printf("serialize proof: [")
	for i, r := range res {
		if i != len(res) - 1 {
			fmt.Printf("BigNumber.from(%q), ", r.String())
		} else {
			fmt.Printf("BigNumber.from(%q)", r.String())
		}
	}
	fmt.Printf("]\n")
	fmt.Println("================================================")
}
