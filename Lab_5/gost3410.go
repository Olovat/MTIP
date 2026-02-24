// Package main реализует ГОСТ Р 34.10-2018 — формирование и проверку цифровой подписи

package main

import (
	"crypto/rand"
	"errors"
	"math/big"
)

//  Вспомогательные арифметические функции по модулю

func sqmod(a, mod *big.Int) *big.Int {
	r := new(big.Int).Mul(a, a)
	return r.Mod(r, mod)
}

func mulmod(a, b, mod *big.Int) *big.Int {
	r := new(big.Int).Mul(a, b)
	return r.Mod(r, mod)
}

func submod(a, b, mod *big.Int) *big.Int {
	r := new(big.Int).Sub(a, b)
	r.Mod(r, mod)
	if r.Sign() < 0 {
		r.Add(r, mod)
	}
	return r
}

func addmod(a, b, mod *big.Int) *big.Int {
	r := new(big.Int).Add(a, b)
	return r.Mod(r, mod)
}

var (
	bigZero  = big.NewInt(0)
	bigOne   = big.NewInt(1)
	bigTwo   = big.NewInt(2)
	bigThree = big.NewInt(3)
	bigEight = big.NewInt(8)
)

//  Параметры эллиптической кривой

// CurveParams содержит параметры кривой y² ≡ x³ + a·x + b (mod p).
type CurveParams struct {
	P  *big.Int
	A  *big.Int
	B  *big.Int
	Q  *big.Int
	Gx *big.Int
	Gy *big.Int
}

func fromHex(s string) *big.Int {
	n, ok := new(big.Int).SetString(s, 16)
	if !ok {
		panic("invalid hex: " + s)
	}
	return n
}

// Curve512Test — тестовый 512-битный набор параметров.
var Curve512Test = &CurveParams{
	P: fromHex("4531ACD1FE0023C7550D267B6B2FEE80922B14B2FFB90F04D4EB7C09" +
		"B5D2D15DF1D852741AF4704A0458047E80E4546D35B8336FAC224DD81664BBF528BE6373"),
	A: fromHex("7"),
	B: fromHex("1CFF0806A31116DA29D8CFA54E57EB748BC5F377E49400FDD788B649" +
		"ECA1AC4361834013B2AD7322480A89CA58E0CF74BC9E540C2ADD6897FAD0A3084F302ADC"),
	Q: fromHex("4531ACD1FE0023C7550D267B6B2FEE80922B14B2FFB90F04D4EB7C09" +
		"B5D2D15DA82F2D7ECB1DBAC719905C5EECC423F1D86E25EDBE23C595D644AAF187E6E6DF"),
	Gx: fromHex("24D19CC64572EE30F396BF6EBBFD7A6C5213B3B3D7057CC825F91093" +
		"A68CD762FD60611262CD838DC6B60AA7EEE804E28BC849977FAC33B4B530F1B120248A9A"),
	Gy: fromHex("2BB312A43BD2CE6E0D020613C857ACDDCFBF061E91E5F2C3F32447C2" +
		"59F39B2C83AB156D77F1496BF7EB3351E1EE4E43DC1A18B91B24640B6DBB92CB1ADD371E"),
}

type jPoint struct {
	X, Y, Z *big.Int
	inf     bool
}

func infinityPoint() *jPoint {
	return &jPoint{X: bigOne, Y: bigOne, Z: bigZero, inf: true}
}

func affineToJacobi(x, y *big.Int) *jPoint {
	return &jPoint{
		X: new(big.Int).Set(x),
		Y: new(big.Int).Set(y),
		Z: big.NewInt(1),
	}
}

// toAffine переводит точку из координат Якоби в аффинные.
func (c *CurveParams) toAffine(p *jPoint) (x, y *big.Int, ok bool) {
	if p.inf || p.Z.Sign() == 0 {
		return nil, nil, false
	}
	pm := c.P
	zInv := new(big.Int).ModInverse(p.Z, pm)
	zInv2 := sqmod(zInv, pm)
	zInv3 := mulmod(zInv2, zInv, pm)
	x = mulmod(p.X, zInv2, pm)
	y = mulmod(p.Y, zInv3, pm)
	return x, y, true
}

// double вычисляет удвоение точки в координатах Якоби.
func (c *CurveParams) double(p1 *jPoint) *jPoint {
	if p1.inf {
		return infinityPoint()
	}
	if p1.Y.Sign() == 0 {
		return infinityPoint()
	}
	pm := c.P

	A := sqmod(p1.X, pm)
	B := sqmod(p1.Y, pm)
	C := sqmod(B, pm)
	t := sqmod(addmod(p1.X, B, pm), pm)
	t = submod(t, A, pm)
	t = submod(t, C, pm)
	D := mulmod(bigTwo, t, pm)
	Z14 := sqmod(sqmod(p1.Z, pm), pm)
	E := addmod(mulmod(bigThree, A, pm), mulmod(c.A, Z14, pm), pm)
	F := sqmod(E, pm)
	X3 := submod(F, mulmod(bigTwo, D, pm), pm)
	Y3 := submod(mulmod(E, submod(D, X3, pm), pm), mulmod(bigEight, C, pm), pm)
	Z3 := mulmod(mulmod(bigTwo, p1.Y, pm), p1.Z, pm)

	return &jPoint{X: X3, Y: Y3, Z: Z3}
}

func (c *CurveParams) mixedAdd(p1 *jPoint, x2, y2 *big.Int) *jPoint {
	if p1.inf {
		return affineToJacobi(x2, y2)
	}
	pm := c.P

	Z1sq := sqmod(p1.Z, pm)
	Z1cb := mulmod(Z1sq, p1.Z, pm)
	U2 := mulmod(x2, Z1sq, pm)
	S2 := mulmod(y2, Z1cb, pm)
	H := submod(U2, p1.X, pm)
	R := submod(S2, p1.Y, pm)

	if H.Sign() == 0 {
		if R.Sign() == 0 {
			return c.double(p1)
		}
		return infinityPoint()
	}

	H2 := sqmod(H, pm)
	H3 := mulmod(H2, H, pm)
	U1H2 := mulmod(p1.X, H2, pm)

	X3 := submod(submod(sqmod(R, pm), H3, pm), mulmod(bigTwo, U1H2, pm), pm)
	Y3 := submod(mulmod(R, submod(U1H2, X3, pm), pm), mulmod(p1.Y, H3, pm), pm)
	Z3 := mulmod(H, p1.Z, pm)

	return &jPoint{X: X3, Y: Y3, Z: Z3}
}

// jacobiAdd складывает две точки Якоби (общий случай).
func (c *CurveParams) jacobiAdd(p1, p2 *jPoint) *jPoint {
	if p1.inf {
		return p2
	}
	if p2.inf {
		return p1
	}
	pm := c.P

	Z1sq := sqmod(p1.Z, pm)
	Z2sq := sqmod(p2.Z, pm)
	U1 := mulmod(p1.X, Z2sq, pm)
	U2 := mulmod(p2.X, Z1sq, pm)
	S1 := mulmod(mulmod(p1.Y, Z2sq, pm), p2.Z, pm)
	S2 := mulmod(mulmod(p2.Y, Z1sq, pm), p1.Z, pm)
	H := submod(U2, U1, pm)
	R := submod(S2, S1, pm)

	if H.Sign() == 0 {
		if R.Sign() == 0 {
			return c.double(p1)
		}
		return infinityPoint()
	}

	H2 := sqmod(H, pm)
	H3 := mulmod(H2, H, pm)
	U1H2 := mulmod(U1, H2, pm)

	X3 := submod(submod(sqmod(R, pm), H3, pm), mulmod(bigTwo, U1H2, pm), pm)
	Y3 := submod(mulmod(R, submod(U1H2, X3, pm), pm), mulmod(S1, H3, pm), pm)
	Z3 := mulmod(mulmod(H, p1.Z, pm), p2.Z, pm)

	return &jPoint{X: X3, Y: Y3, Z: Z3}
}

// scalarMult вычисляет k·(px, py) с использованием смешанного сложения.
func (c *CurveParams) scalarMult(k *big.Int, px, py *big.Int) *jPoint {
	result := infinityPoint()
	for i := k.BitLen() - 1; i >= 0; i-- {
		result = c.double(result)
		if k.Bit(i) == 1 {
			result = c.mixedAdd(result, px, py)
		}
	}
	return result
}

// AffinePoint — точка эллиптической кривой в аффинных координатах
type AffinePoint struct {
	X, Y *big.Int
}

// PrivateKey — закрытый ключ.
type PrivateKey struct {
	D     *big.Int // закрытый ключ
	Curve *CurveParams
}

// PublicKey — открытый ключ
type PublicKey struct {
	Q     *AffinePoint //точка на кривой
	Curve *CurveParams
}

// KeyPair — пара ключей
type KeyPair struct {
	Private PrivateKey
	Public  PublicKey
}

//  Генерация ключей

// GenerateKeyPair генерирует пару ключей для ГОСТ Р 34.10-2018.
//

func GenerateKeyPair(curve *CurveParams) (KeyPair, error) {
	qMinus1 := new(big.Int).Sub(curve.Q, bigOne)
	d, err := rand.Int(rand.Reader, qMinus1)
	if err != nil {
		return KeyPair{}, err
	}
	d.Add(d, bigOne) // d ∈ [1, q-1]

	Qj := curve.scalarMult(d, curve.Gx, curve.Gy)
	qx, qy, ok := curve.toAffine(Qj)
	if !ok {
		return KeyPair{}, errors.New("GenerateKeyPair: d·G = ∞ (curve error)")
	}

	return KeyPair{
		Private: PrivateKey{D: d, Curve: curve},
		Public:  PublicKey{Q: &AffinePoint{X: qx, Y: qy}, Curve: curve},
	}, nil
}

//  Формирование подписи — ГОСТ Р 34.10-2018, Алгоритм I

func Sign(msg []byte, priv PrivateKey) (r, s *big.Int, err error) {
	curve := priv.Curve

	// Шаги 1–2: хэш и вычисление e
	hBytes := streebog512(msg)
	alpha := new(big.Int).SetBytes(hBytes)
	e := new(big.Int).Mod(alpha, curve.Q)
	if e.Sign() == 0 {
		e.SetInt64(1)
	}

	qMinus1 := new(big.Int).Sub(curve.Q, bigOne)

	for {
		// Шаг 3: случайное k
		k, er := rand.Int(rand.Reader, qMinus1)
		if er != nil {
			err = er
			return
		}
		k.Add(k, bigOne)

		// Шаг 4: C = k·G;  r = x_C mod q
		Cj := curve.scalarMult(k, curve.Gx, curve.Gy)
		cx, _, ok := curve.toAffine(Cj)
		if !ok {
			continue
		}
		r = new(big.Int).Mod(cx, curve.Q)
		if r.Sign() == 0 {
			continue
		}

		// Шаг 5: s = (r·d + k·e) mod q
		rd := mulmod(r, priv.D, curve.Q)
		ke := mulmod(k, e, curve.Q)
		s = addmod(rd, ke, curve.Q)
		if s.Sign() == 0 {
			continue
		}

		return r, s, nil
	}
}

//  Проверка подписи — ГОСТ Р 34.10-2018, Алгоритм II

// Verify проверяет ЭЦП (r, s) сообщения msg открытым ключом pub.

func Verify(msg []byte, r, s *big.Int, pub PublicKey) bool {
	curve := pub.Curve

	// Шаг 1: диапазон r и s
	if r.Sign() <= 0 || r.Cmp(curve.Q) >= 0 {
		return false
	}
	if s.Sign() <= 0 || s.Cmp(curve.Q) >= 0 {
		return false
	}

	// Шаги 2–3: хэш и вычисление e
	hBytes := streebog512(msg)
	alpha := new(big.Int).SetBytes(hBytes)
	e := new(big.Int).Mod(alpha, curve.Q)
	if e.Sign() == 0 {
		e.SetInt64(1)
	}

	v := new(big.Int).ModInverse(e, curve.Q)

	z1 := mulmod(s, v, curve.Q)
	z2 := mulmod(r, v, curve.Q)
	z2 = new(big.Int).Sub(curve.Q, z2)

	C1 := curve.scalarMult(z1, curve.Gx, curve.Gy)
	C2 := curve.scalarMult(z2, pub.Q.X, pub.Q.Y)
	Cj := curve.jacobiAdd(C1, C2)
	cx, _, ok := curve.toAffine(Cj)
	if !ok {
		return false
	}

	// подпись верна, если  R == r
	R := new(big.Int).Mod(cx, curve.Q)
	return R.Cmp(r) == 0
}
