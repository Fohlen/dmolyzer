package game

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
