package dto

import "backend-election/internal/model"

type PemilihResponse struct {
	PemilihID    string `json:"pemilih_id"`
	Nama         string `json:"nama"`
	NIK          string `json:"nik"`
	TanggalLahir string `json:"tanggal_lahir"`
	Alamat       string `json:"alamat"`
	SidikJari    string `json:"sidik_jari"`
}

func (u *PemilihResponse) FromEntity(pemilih model.Pemilih) {
	u.PemilihID = pemilih.PemilihID
	u.Nama = pemilih.Nama
	u.NIK = pemilih.NIK
	u.TanggalLahir = pemilih.TanggalLahir
	u.Alamat = pemilih.Alamat
	u.SidikJari = pemilih.SidikJari
}
