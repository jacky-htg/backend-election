package repository

import (
	"backend-election/internal/model"
	"backend-election/internal/pkg/logger"
	"context"
	"database/sql"
)

type PemilihRepository struct {
	Db            *sql.DB
	Log           *logger.Logger
	PemilihEntity model.Pemilih
}

func (u *PemilihRepository) FindAll(ctx context.Context) ([]model.Pemilih, error) {
	// Check if context is already canceled or deadline exceeded
	if ctx.Err() != nil {
		return nil, u.Log.Error(ctx.Err())
	}

	const q = `SELECT pemilih_id, nik, nama, tanggal_lahir, alamat, sidik_jari FROM pemilih`
	rows, err := u.Db.QueryContext(ctx, q)
	if err != nil {
		return nil, u.Log.Error(err)
	}
	defer rows.Close()

	var pemilihList []model.Pemilih
	for rows.Next() {
		var pemilih model.Pemilih
		err := rows.Scan(&pemilih.PemilihID, &pemilih.NIK, &pemilih.Nama, &pemilih.TanggalLahir, &pemilih.Alamat, &pemilih.SidikJari)
		if err != nil {
			return nil, u.Log.Error(err)
		}
		pemilihList = append(pemilihList, pemilih)
	}

	if err = rows.Err(); err != nil {
		return nil, u.Log.Error(err)
	}

	return pemilihList, nil
}

func (u *PemilihRepository) Find(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return u.Log.Error(context.Canceled)
	case context.DeadlineExceeded:
		return u.Log.Error(context.DeadlineExceeded)
	default:
	}

	const q = `SELECT pemilih_id, nik, nama, tanggal_lahir, alamat, sidik_jari FROM pemilih WHERE pemilih_id=$1`
	stmt, err := u.Db.PrepareContext(ctx, q)
	if err != nil {
		return u.Log.Error(err)
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, u.PemilihEntity.PemilihID).Scan(&u.PemilihEntity.PemilihID, &u.PemilihEntity.NIK, &u.PemilihEntity.Nama, &u.PemilihEntity.TanggalLahir, &u.PemilihEntity.Alamat, &u.PemilihEntity.SidikJari)
	if err != nil {
		return u.Log.Error(err)
	}
	return nil
}
