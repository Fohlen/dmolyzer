package main

import (
	"bytes"
	"compress/flate"
	"dmolyzer/pkg/game"
	"dmolyzer/pkg/packet"
	"dmolyzer/pkg/parser"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var headers = []string{"Time", "Event", "Name", "Team", "Value"}

func writePlayerLog(currentTime int, of io.Writer, g *game.Game) {
	for _, p := range g.Players {
		if !p.Connected || (p.State < 0 || p.State > 4) || p.DamageDealt == 0 {
			continue
		}

		fmt.Fprintf(of, "%d\tPosition\t%s\t%s", currentTime, p.Name, p.Team)
		fmt.Fprintf(of, "\t%d,%d,%d", p.Position.X, p.Position.Y, p.Position.Y)
		fmt.Fprintln(of)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ./recordings file|directory [file|directory [...]]")
	}

	for _, arg := range os.Args[1:] {
		info, err := os.Stat(arg)
		if err != nil {
			log.Fatal(err)
		}

		files := make([]string, 0)
		fpath, fname := filepath.Split(arg)
		var outFileName string

		if info.IsDir() {
			fileinfos, err2 := ioutil.ReadDir(arg)
			if err2 != nil {
				log.Fatal(err2)
			}

			for _, fileinfo := range fileinfos {
				files = append(files, filepath.Join(arg, fileinfo.Name()))
			}

			outFileName = filepath.Join(fpath, fname+".tsv")
		} else {
			files = append(files, filepath.Join(fpath, arg))
			outFileName = filepath.Join(fpath, strings.TrimSuffix(fname, filepath.Ext(fname))+".tsv")
		}

		of, err := os.Create(outFileName)
		if err != nil {
			log.Fatal(err)
		}
		defer of.Close()

		fmt.Fprintf(of, "%s\n", strings.Join(headers, "\t"))

		for _, file := range files {
			info, err = os.Stat(file)
			if err != nil {
				log.Fatal(err)
			}

			g := game.Game{}
			g.Time = info.ModTime().UnixNano() / int64(time.Second)
			g.Players = make([]game.Player, 128)

			f, err := os.Open(file)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			fileHeader, _ := packet.ReadNextBytes(f, 10)

			if bytes.Compare(parser.ZHeader, fileHeader) != 0 {
				log.Fatal("invalid zlib header in", file)
			}

			fz := flate.NewReader(f)
			if err != nil {
				log.Fatal(err)
			}
			defer fz.Close()

			zDemoHeader, _ := packet.ReadNextBytes(fz, 24)

			if bytes.Compare(parser.DemoHeader, zDemoHeader) != 0 {
				log.Fatal("invalid demo header in", file)
			}

			var lastTime int

			for {
				cTime, err := packet.ReadNextBytes(fz, 4)
				if err == io.EOF {
					break
				}
				ch, _ := packet.ReadNextBytes(fz, 4)
				len, _ := packet.ReadNextBytes(fz, 4)
				data, _ := packet.ReadNextBytes(fz, int(binary.LittleEndian.Uint32(len)))

				g.CurTime = int(binary.LittleEndian.Uint32(cTime))
				if lastTime == 0 {
					lastTime = g.CurTime
				}

				if int(binary.LittleEndian.Uint32(ch)) == 0 {
					parser.ParsePositions(&data, &g)
				} else if int(binary.LittleEndian.Uint32(ch)) == 1 {
					parser.ParseMessage(&data, &g)
				}

				if g.CurTime > (lastTime + 1000) {
					lastTime = g.CurTime
					writePlayerLog(lastTime, of, &g)
				}

				if g.CurTime >= g.EndTime {
					break
				}
			}

		}
	}
}
