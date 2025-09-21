package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User of the system
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Username     string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	Email        string    `gorm:"uniqueIndex"`
	UsedStorage  int64     `gorm:"default:0"`        // bytes used
	Quota        int64     `gorm:"default:10485760"` // default 10 MB (configurable)
	IsAdmin      bool      `gorm:"default:false"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`

	UserFiles []UserFile
	Folders   []Folder
}

// File represents unique deduplicated content
type File struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Hash       string    `gorm:"uniqueIndex;not null"` // sha256 hex
	ObjectName string    `gorm:"not null"`             // object key in MinIO
	Size       int64     `gorm:"not null"`
	MimeType   string
	RefCount   int       `gorm:"default:1"` // number of users referencing used for deduplication catch
	CreatedAt  time.Time `gorm:"autoCreateTime"`

	UserFiles []UserFile
}

// UserFile links a user to a file, with sharing metadata
type UserFile struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	FileID     uuid.UUID `gorm:"type:uuid;not null;index"`
	IsOwner    bool      `gorm:"default:false"`
	Visibility string    `gorm:"type:text;default:'private'"` // private | public | shared
	Downloads  int64     `gorm:"default:0"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	File File `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE"`
}

// Folder
type Folder struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string
	ParentID  *uuid.UUID `gorm:"type:uuid"`
	OwnerID   uuid.UUID  `gorm:"type:uuid"`
	Owner     User
	Children  []Folder   `gorm:"foreignKey:ParentID"`
	Files     []UserFile `gorm:"foreignKey:FolderID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Share struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FileID     uuid.UUID `gorm:"type:uuid;not null"`
	UserID     uuid.UUID `gorm:"type:uuid;not null"`
	IsPublic   bool      `gorm:"default:false"`
	SharedWith *string   //optional
	Downloads  int       `gorm:"default:0"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`

	File File `gorm:"foreignKey:FileID"`
	User User `gorm:"foreignKEy:UserID"`
}

// BeforeCreate hooks to auto-generate UUIDs if Postgres function gen_random_uuid() isn’t available
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
func (f *File) BeforeCreate(tx *gorm.DB) (err error) {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return
}
func (uf *UserFile) BeforeCreate(tx *gorm.DB) (err error) {
	if uf.ID == uuid.Nil {
		uf.ID = uuid.New()
	}
	return
}
func (fo *Folder) BeforeCreate(tx *gorm.DB) (err error) {
	if fo.ID == uuid.Nil {
		fo.ID = uuid.New()
	}
	return
}
