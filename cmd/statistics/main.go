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

var headers = []string{"Time", "Mode", "Map", "Name", "Team", "Frags", "Deaths", "Damage", "DamageDealt",
	"Suicides", "TotalShots", "ShotsDealt", "FistDamage", "FistDamageDealt", "ShotgunDamage", "ShotgunDamageDealt", "ChaingunDamage", "ChaingunDamageDealt",
	"RocketLauncherDamage", "RocketLauncherDamageDealt", "RifleDamage", "RifleDamageDealt", "GrenadeLauncherDamage", "GrenadeLauncherDamageDealt",
	"PistolDamage", "PistolDamageDealt", "FistShots", "FistShotsDealt", "ShotgunShots", "ShotgunShotsDealt", "ChaingunShots", "ChaingunShotsDealt",
	"RocketLauncherShots", "RocketLauncherShotsDealt", "RifleShots", "RifleShotsDealt", "GrenadeLauncherShots", "GrenadeLauncherShotsDealt", "PistolShots", "PistolShotsDealt",
	"FlagsScored", "FlagsResetted", "FlagsDropped"}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ./statistics file|directory [file|directory [...]]")
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

			fname = filepath.Base(fpath)
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

			for {
				cTime, err := packet.ReadNextBytes(fz, 4)
				if err == io.EOF {
					break
				}
				ch, _ := packet.ReadNextBytes(fz, 4)
				len, _ := packet.ReadNextBytes(fz, 4)
				data, _ := packet.ReadNextBytes(fz, int(binary.LittleEndian.Uint32(len)))

				g.CurTime = int(binary.LittleEndian.Uint32(cTime))

				if int(binary.LittleEndian.Uint32(ch)) == 1 {
					parser.ParseMessage(&data, &g)
				}

				if g.CurTime >= g.EndTime {
					break
				}
			}

			for _, p := range g.Players {
				if !p.Connected || (p.State >= 4) {
					continue
				}

				fmt.Fprintf(of, "%d\t%d\t%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%d", g.Time, g.Mode, g.Map, p.Name, p.Team, p.Frags, p.Deaths, p.Damage, p.DamageDealt, p.Suicides, p.TotalShots, p.ShotsDealt)
				for i := 0; i < 7; i++ {
					fmt.Fprintf(of, "\t%d\t%d", p.WeaponDamage[i], p.WeaponDamageDealt[i])
				}
				for i := 0; i < 7; i++ {
					fmt.Fprintf(of, "\t%d\t%d", p.WeaponShots[i], p.WeaponShotsDealt[i])
				}
				fmt.Fprintf(of, "\t%d\t%d\t%d", p.FlagsScored, p.FlagsResetted, p.FlagsDropped)
				fmt.Fprintln(of)
			}
		}
	}
}
