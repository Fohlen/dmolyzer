package dmolyzer

import (
	"encoding/binary"
	"io"
)

var gunsDamage = []int{50, 200, 30, 120, 100, 90, 35}
var cube2UniCars = []rune{0, 192, 193, 194, 195, 196, 197, 198, 199, 9, 10, 11, 12, 13, 200, 201,
	202, 203, 204, 205, 206, 207, 209, 210, 211, 212, 213, 214, 216, 217, 218, 219,
	32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47,
	48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63,
	64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79,
	80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95,
	96, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111,
	112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 220,
	221, 223, 224, 225, 226, 227, 228, 229, 230, 231, 232, 233, 234, 235, 236, 237,
	238, 239, 241, 242, 243, 244, 245, 246, 248, 249, 250, 251, 252, 253, 255, 0x104,
	0x105, 0x106, 0x107, 0x10C, 0x10D, 0x10E, 0x10F, 0x118, 0x119, 0x11A, 0x11B, 0x11E, 0x11F, 0x130, 0x131, 0x141,
	0x142, 0x143, 0x144, 0x147, 0x148, 0x150, 0x151, 0x152, 0x153, 0x158, 0x159, 0x15A, 0x15B, 0x15E, 0x15F, 0x160,
	0x161, 0x164, 0x165, 0x16E, 0x16F, 0x170, 0x171, 0x178, 0x179, 0x17A, 0x17B, 0x17C, 0x17D, 0x17E, 0x404, 0x411,
	0x413, 0x414, 0x416, 0x417, 0x418, 0x419, 0x41B, 0x41F, 0x423, 0x424, 0x426, 0x427, 0x428, 0x429, 0x42A, 0x42B,
	0x42C, 0x42D, 0x42E, 0x42F, 0x431, 0x432, 0x433, 0x434, 0x436, 0x437, 0x438, 0x439, 0x43A, 0x43B, 0x43C, 0x43D,
	0x43F, 0x442, 0x444, 0x446, 0x447, 0x448, 0x449, 0x44A, 0x44B, 0x44C, 0x44D, 0x44E, 0x44F, 0x454, 0x490, 0x491}

type Packet struct {
	Data *[]byte
	Pos  int
}

func (p *Packet) GetBytes(length int) []byte {
	buf := (*p.Data)[p.Pos:(p.Pos + length)]
	p.Pos += length
	return buf
}

func (p *Packet) GetByte() byte {
	b := (*p.Data)[p.Pos]
	p.Pos++
	return b
}

func (p *Packet) GetInt() int {
	n := int8((*p.Data)[p.Pos])
	result := int(n)
	if n == -128 {
		result = int((*p.Data)[p.Pos+1]) | int((*p.Data)[p.Pos+2])<<8
		p.Pos += 2
	} else if n == -127 {
		result = int((*p.Data)[p.Pos+1]) | int((*p.Data)[p.Pos+2])<<8 | int((*p.Data)[p.Pos+3])<<16 | int((*p.Data)[p.Pos+4])<<24
		p.Pos += 4
	}
	p.Pos++
	return result
}

func (p *Packet) GetString() string {
	result := ""
	for ; p.Pos < len(*p.Data); p.Pos++ {
		if (*p.Data)[p.Pos] == '\x00' {
			p.Pos++
			break
		}
		result += string(cube2UniCars[(*p.Data)[p.Pos]])
	}
	return result
}

type Game struct {
	Time    int64
	Mode    int
	Map     string
	Players []Player
	EndTime int
	CurTime int
}

type PlayerPosition struct {
	X int
	Y int
	Z int
}

type Player struct {
	Name              string
	Team              string
	Position          PlayerPosition
	Frags             int
	Deaths            int
	Damage            int
	DamageDealt       int
	Suicides          int
	WeaponDamage      [7]int
	WeaponDamageDealt [7]int
	LastWeapon        int
	TotalShots        int
	ShotsDealt        int
	WeaponShots       [7]int
	WeaponShotsDealt  [7]int
	FlagsScored       int
	FlagsDropped      int
	FlagsResetted     int
	Connected         bool
	State             int
	Model             int
}

func ReadNextBytes(file io.Reader, number int) ([]byte, error) {
	bytes := make([]byte, number)

	n, err := file.Read(bytes)
	if err != nil {
		if err == io.EOF {
			return bytes, err
		}
		return nil, err
	}

	if n != number {
		newBytes, err := ReadNextBytes(file, number-n)
		copy(bytes[n:], newBytes)
		if err != nil {
			return bytes, err
		}
	}

	return bytes, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func ParsePositions(msg *[]byte, g *Game) {
	p := Packet{msg, 0}
	msgType := p.GetInt()

	if msgType == 4 { // N_POS
		// NOTE: See a documentation of N_POS and potential things you could include here
		// https://kb.p1x.pw/np-position-packet/
		cn := p.GetInt()
		p.GetByte()
		flags := p.GetByte()
		x3 := (flags & (1 << 0)) != 0
		y3 := (flags & (1 << 1)) != 0
		z3 := (flags & (1 << 2)) != 0
		// Unlike for other messages we may have 3-byte integers
		x := binary.LittleEndian.Uint16(p.GetBytes(2 + boolToInt(x3)))
		y := binary.LittleEndian.Uint16(p.GetBytes(2 + boolToInt(y3)))
		z := binary.LittleEndian.Uint16(p.GetBytes(2 + boolToInt(z3)))
		g.Players[cn].Position.X = int(x)
		g.Players[cn].Position.Y = int(y)
		g.Players[cn].Position.Y = int(z)
		// Discard the remaining information
		p.Pos = cap(*p.Data)
	}
}

func ParseMessage(msg *[]byte, g *Game) {
	p := Packet{msg, 0}
	msgType := p.GetInt()

	switch msgType {
	case 2: // N_WELCOME
		for p.Pos < len(*p.Data)-1 {
			t := p.GetInt()
			switch t {
			case 22: // N_MAPCHANGE
				g.Map = p.GetString()
				g.Mode = p.GetInt()
				p.GetInt()
			case 33: // N_TIMEUP
				timeLeft := p.GetInt()
				g.EndTime = g.CurTime + timeLeft*1000
			case 36: // N_ITEMLIST
				for p.GetInt() != -1 {
					p.GetInt()
				}
			case 58: // N_CURRENTMASTER
				p.GetInt()
				for p.GetInt() != -1 {
					p.GetInt()
				}
			case 91: // N_PAUSEGAME
				p.GetInt()
				p.GetInt()
			case 92: // N_GAMESPEED
				p.GetInt()
				p.GetInt()
			case 24: // N_TEAMINFO
				for p.GetString() != "" {
					p.GetInt()
				}
			case 61: // N_SETTEAM
				p.GetInt()
				p.GetString()
				p.GetInt()
			case 19: // N_FORCEDEATH
				p.GetInt()
			case 17: // N_SPAWNSTATE
				for i := 0; i < 13; i++ {
					p.GetInt()
				}
			case 59: // N_SPECTATOR
				p.GetInt()
				p.GetInt()
			case 37: // N_RESUME
				for cn := p.GetInt(); cn != -1; cn = p.GetInt() {
					g.Players[cn].State = p.GetInt()
					g.Players[cn].Frags = p.GetInt()
					for i := 0; i < 15; i++ {
						p.GetInt()
					}
				}
			case 95: // N_INITAI
				for i := 0; i < 5; i++ {
					p.GetInt()
				}
				p.GetString()
				p.GetString()
			case 3: // N_INITCLIENT
				cn := p.GetInt()
				name := p.GetString()
				team := p.GetString()
				model := p.GetInt()
				g.Players[cn].Name = name
				g.Players[cn].Team = team
				g.Players[cn].Model = model
				g.Players[cn].Connected = true
			}
		}
	case 3: // N_INITCLIENT
		cn := p.GetInt()
		name := p.GetString()
		g.Players[cn].Name = name
		g.Players[cn].Connected = true
		g.Players[cn].State = -1
	case 59: // N_SPECTATOR
		cn := p.GetInt()
		spectator := p.GetInt()
		if spectator != 0 {
			g.Players[cn].State = 5
		} else {
			g.Players[cn].State = 1
		}
	case 7: // N_CDIS
		cn := p.GetInt()
		g.Players[cn].Connected = false
		g.Players[cn].State = 5
	case 15: // N_EXPLODEFX
		attacker := p.GetInt()
		gun := p.GetInt()
		if gun != 3 && gun != 5 {
			break
		}
		g.Players[attacker].LastWeapon = gun
	case 14: // N_SHOTFX
		attacker := p.GetInt()
		gun := p.GetInt()
		if gun < 0 || gun > 6 {
			break
		}
		g.Players[attacker].LastWeapon = gun
		g.Players[attacker].Damage += gunsDamage[gun]
		g.Players[attacker].WeaponDamage[gun] += gunsDamage[gun]
		g.Players[attacker].WeaponShots[gun] += 1
		g.Players[attacker].TotalShots += 1
	case 12: // N_DAMAGE
		victim := p.GetInt()
		attacker := p.GetInt()
		damage := p.GetInt()
		if attacker != victim {
			g.Players[attacker].DamageDealt += damage
			g.Players[attacker].ShotsDealt += 1
		}
		gun := g.Players[attacker].LastWeapon
		g.Players[attacker].WeaponDamageDealt[gun] += damage
		g.Players[attacker].WeaponShotsDealt[gun] += 1
	case 11: // N_DIED
		victim := p.GetInt()
		attacker := p.GetInt()
		frags := p.GetInt()
		g.Players[attacker].Frags = frags
		g.Players[victim].Deaths++
		if victim == attacker {
			g.Players[attacker].Suicides++
		}
	case 79: // N_TAKEFLAG
	case 80: // N_RETURNFLAG
		cn := p.GetInt()
		p.GetInt()
		p.GetInt()
		g.Players[cn].FlagsResetted += 1
	case 84: // N_DROPFLAG
		cn := p.GetInt()
		for i := 0; i < 4; i++ {
			// NOTE: We can get positions here
			p.GetInt()
		}
		g.Players[cn].FlagsDropped += 1
	case 85: // N_SCOREFLAG
		ocn := p.GetInt()
		p.GetInt() // relayflag
		p.GetInt() // relayversion
		p.GetInt() // goalflag
		p.GetInt() // goalversion
		p.GetInt() // goalspawn
		p.GetInt() // team
		p.GetInt() // score
		p.GetInt() // oflags
		g.Players[ocn].FlagsScored += 1
	}
}
