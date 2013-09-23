package models

import (
	"encoding/json"
	"github.com/songgao/squirrel/models/common"
	"net/http"
	"os/exec"
	"path"
	"strings"
)

type interactivePositions struct {
	positionManager *common.PositionManager
	newPositions    chan *common.Position
	laddr           string
}

func newInteractivePositions() common.MobilityManager {
	return &interactivePositions{newPositions: make(chan *common.Position)}
}

func (m *interactivePositions) ParametersHelp() string {
	return ``
}

func (m *interactivePositions) Configure(config map[string]interface{}) error {
	var ok bool
	m.laddr, ok = config["laddr"].(string)
	if !ok {
		return ParametersNotValid
	}
	return nil
}

func (m *interactivePositions) Initialize(positionManager *common.PositionManager) {
	m.positionManager = positionManager
	go http.ListenAndServe(m.laddr, m.bindMux())
}

type JSPosition struct {
	I int
	X float64
	Y float64
	H float64
}

func positionFromPosition(i int, p *common.Position) *JSPosition {
	return &JSPosition{I: i, X: p.X, Y: p.Y, H: p.Height}
}

func (m *interactivePositions) bindMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/list", func(w http.ResponseWriter, req *http.Request) {
		ret := make([]*JSPosition, 0)
		for _, index := range m.positionManager.Enabled() {
			p := m.positionManager.Get(index)
			ret = append(ret, positionFromPosition(index, &p))
		}
		json.NewEncoder(w).Encode(ret)
	})
	mux.HandleFunc("/set", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			http.NotFound(w, req)
			return
		}
		var pos JSPosition
		err := json.NewDecoder(req.Body).Decode(&pos)
		if nil != err {
			http.Error(w, "json Decoding error", 500)
			return
		}
		m.positionManager.Set(pos.I, pos.X, pos.Y, pos.H)
	})
	pkgRoot, err := getRootPath()
	if err != nil {
		return nil
	}
	mux.Handle("/", http.FileServer(http.Dir(path.Join(pkgRoot, "interactivePositions"))))

	return mux
}

func getRootPath() (string, error) {
	out, err := exec.Command("go", "list", "-f", "{{.Dir}}", "github.com/songgao/squirrel/models").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
