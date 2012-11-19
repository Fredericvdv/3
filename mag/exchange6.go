package mag

import (
	"code.google.com/p/nimble-cube/core"
	"code.google.com/p/nimble-cube/nimble"
)

type Exchange6 struct {
	m           nimble.RChan3
	hex         nimble.Chan3
	aex_reduced float64
}

func (e *Exchange6) Run() {
	N := e.m.Mesh().NCell()
	size := e.m.Mesh().Size()
	cellsize := e.m.Mesh().CellSize()
	// TODO: properly split in blocks
	for {
		mSlice := e.m.ReadNext(N)
		mList := [3][]float32{mSlice[0].Host(), mSlice[1].Host(), mSlice[2].Host()}
		m := core.Reshape3(mList, size)
		hex := core.Reshape3(e.hex.WriteNext(N), size)
		exchange6(m, hex, cellsize, e.aex_reduced)
		e.hex.WriteDone()
		e.m.ReadDone()
	}
}

func NewExchange6(tag, unit string, memType nimble.MemType, m nimble.RChan3, aex_reduced float64) *Exchange6 {
	hex := nimble.MakeChan3(tag, unit, m.Mesh(), memType, 0)
	return &Exchange6{m, hex, aex_reduced}
}

func (e *Exchange6) Output() nimble.Chan3 {
	return e.hex
}

// Naive implementation of 6-neighbor exchange field.
// Aex in Tm² (exchange stiffness divided by Msat0).
// Hex in Tesla.
func exchange6(m [3][][][]float32, Hex [3][][][]float32, cellsize [3]float64, aex_reduced float64) {
	var (
		facI = float32(aex_reduced / (cellsize[0] * cellsize[0]))
		facJ = float32(aex_reduced / (cellsize[1] * cellsize[1]))
		facK = float32(aex_reduced / (cellsize[2] * cellsize[2]))
	)
	N0, N1, N2 := len(m[0]), len(m[0][0]), len(m[0][0][0])

	for i := range m[0] {
		// could hack in here (for 3D)...
		for j := range m[0][i] {
			// or here (for 2D) and synchronize
			for k := range m[0][i][j] {
				m0 := Vector{m[X][i][j][k],
					m[Y][i][j][k],
					m[Z][i][j][k]}
				var hex, m_neigh Vector

				if i > 0 {
					m_neigh = Vector{m[X][i-1][j][k],
						m[Y][i-1][j][k],
						m[Z][i-1][j][k]}
					hex = hex.Add((m_neigh.Sub(m0)).Scaled(facI))
				}

				if i < N0-1 {
					m_neigh = Vector{m[X][i+1][j][k],
						m[Y][i+1][j][k],
						m[Z][i+1][j][k]}
					hex = hex.Add((m_neigh.Sub(m0)).Scaled(facI))
				}

				if j > 0 {
					m_neigh = Vector{m[X][i][j-1][k],
						m[Y][i][j-1][k],
						m[Z][i][j-1][k]}
					hex = hex.Add((m_neigh.Sub(m0)).Scaled(facJ))
				}

				if j < N1-1 {
					m_neigh = Vector{m[X][i][j+1][k],
						m[Y][i][j+1][k],
						m[Z][i][j+1][k]}
					hex = hex.Add((m_neigh.Sub(m0)).Scaled(facJ))
				}

				if k > 0 {
					m_neigh = Vector{m[X][i][j][k-1],
						m[Y][i][j][k-1],
						m[Z][i][j][k-1]}
					hex = hex.Add((m_neigh.Sub(m0)).Scaled(facK))
				}

				if k < N2-1 {
					m_neigh = Vector{m[X][i][j][k+1],
						m[Y][i][j][k+1],
						m[Z][i][j][k+1]}
					hex = hex.Add((m_neigh.Sub(m0)).Scaled(facK))
				}

				Hex[X][i][j][k] = hex[X]
				Hex[Y][i][j][k] = hex[Y]
				Hex[Z][i][j][k] = hex[Z]
			}
		}
	}
}
