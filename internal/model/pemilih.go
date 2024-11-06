package model

type Pemilih struct {
	PemilihID    string `db:"pemilih_id"`
	NIK          string `db:"nik"`
	Nama         string `db:"nama"`
	TanggalLahir string `db:"tanggal_lahir"`
	Alamat       string `db:"alamat"`
	SidikJari    string `db:"sidik_jari"`
	StatusAktif  bool   `db:"status_aktif"`
}
