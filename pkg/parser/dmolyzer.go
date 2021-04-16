package parser

import (
	"dmolyzer/internal/pkg/helpers"
	"dmolyzer/pkg/game"
	"dmolyzer/pkg/packet"
	"encoding/binary"
)

var gunsDamage = []int{50, 200, 30, 120, 100, 90, 35}

func ParsePositions(msg *[]byte, g *game.Game) {
	p := packet.Packet{msg, 0}
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
		x := binary.LittleEndian.Uint16(p.GetBytes(2 + helpers.BoolToInt(x3)))
		y := binary.LittleEndian.Uint16(p.GetBytes(2 + helpers.BoolToInt(y3)))
		z := binary.LittleEndian.Uint16(p.GetBytes(2 + helpers.BoolToInt(z3)))
		g.Players[cn].Position.X = int(x)
		g.Players[cn].Position.Y = int(y)
		g.Players[cn].Position.Y = int(z)
		// Discard the remaining information
		p.Pos = cap(*p.Data)
	}
}

func ParseMessage(msg *[]byte, g *game.Game) {
	p := packet.Packet{msg, 0}
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
