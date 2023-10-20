package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/adarsh-shahi/gotube-api/internals/youtube"
)

type PostgreDB struct {
	PDB       *sql.DB
	DbTimeout time.Duration
}

func (pDB *PostgreDB) Connection() *sql.DB {
	return pDB.PDB
}

type TIdEmailPassword struct {
	Id       int64  `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	UType    string `json:"utype"`
	Role     string `json:"role"`
}

func (pDB *PostgreDB) GetIdEmailUser(user TIdEmailPassword) (*TIdEmailPassword, error) {
	ctx, cancel := context.WithTimeout(context.Background(), pDB.DbTimeout)
	defer cancel()

	var table string
	if user.UType == "owner" {
		table = "owners"
	} else {
		table = "users"
	}

	query := fmt.Sprintf("select id, email from %s where email = '%s'", table, user.Email)
	row := pDB.PDB.QueryRowContext(ctx, query)

	u := TIdEmailPassword{}

	err := row.Scan(
		&u.Id,
		&u.Email,
	)
	u.UType = user.UType
	u.Role = user.Role
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, errors.New("user not found")
		default:
			return nil, errors.New("something went wrong")
		}
	}
	return &u, nil
}

type AddUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AddOwner struct {
	youtube.Channel
	Email string `json:"email"`
}

func (pDB *PostgreDB) AddOwner(owner AddOwner) (int64, error) {
	var id int64
	query := fmt.Sprintf(
		"insert into owners(email, channelName, channelUrl, profileImage, description) values ('%s','%s','%s','%s','%s') RETURNING id",
		owner.Email,
		owner.Title,
		owner.CustomUrl,
		owner.ProfileImageUrl,
		owner.Description,
	)
	err := pDB.PDB.QueryRowContext(context.Background(), query).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (pDB *PostgreDB) AddUser(user AddUser) (int64, error) {
	var id int64
	query := fmt.Sprintf("insert into users(email) values('%s') RETURNING id", user.Email)
	err := pDB.PDB.QueryRowContext(context.Background(), query).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (pDB *PostgreDB) AddUserWithPassowrd(user AddUser) error {
	query := fmt.Sprintf("insert into users(email, password) values('%s', '%s')", user.Email, user.Password)
	_, err := pDB.PDB.ExecContext(context.Background(), query)
	if err != nil {
		return err
	}
	return nil
}

func (pDB *PostgreDB) AddInvites(sender, receiver int64, role string) error {
	query := fmt.Sprintf("insert into invites(sender, receiver, role) values(%d, %d, '%s')", sender, receiver, role)
	_, err := pDB.PDB.ExecContext(context.Background(), query)
	if err != nil {
		return err
	}
	return nil
}

func (pDB *PostgreDB) UpdateInviteRole(sender, receiver int64, role string) error {
	query := fmt.Sprintf("update invites set role = '%s' where sender = %d AND receiver = %d", role, sender, receiver)
	_, err := pDB.PDB.ExecContext(context.Background(), query)
	if err != nil {
		return err
	}
	return nil
}

func (pDB *PostgreDB) DeleteInvite(sender, receiver int64) error {
	query := fmt.Sprintf("delete from invites where sender = %d AND receiver = %d", sender, receiver)
	_, err := pDB.PDB.ExecContext(context.Background(), query)
	if err != nil {
		return err
	}
	return nil
}

func (pDB *PostgreDB) DeleteInviteAndAddToTeam(sender, receiver int64, role string) error {
	ctx := context.Background()
	tx, err := pDB.PDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	queryToDelete := fmt.Sprintf(
		"delete from invites where sender = %d and receiver = %d and role = '%s'",
		sender,
		receiver,
		role,
	)
	result, err := tx.ExecContext(ctx, queryToDelete)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if rows != 1 {
		tx.Rollback()
		return errors.New("unable to find invite")
	}

	queryToAdd := fmt.Sprintf("insert into teams(owner, member, role) values(%d, %d, '%s')", sender, receiver, role)
	result, err = tx.ExecContext(ctx, queryToAdd)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

type Invites struct {
	Email       string `json:"email"`
	Role        string `json:"role"`
	ChannelName string `json:"channelName"`
}

func (pDB *PostgreDB) GetAllInvites(id int64, utype string) ([]Invites, error) {
	var list []Invites
	query := ""

	if utype == "owner" {
		// get all invites sent by the owner to his users (team)
		query = fmt.Sprintf(
			"select users.email, invites.role from invites inner join users on invites.receiver = users.id where invites.sender = %d",
			id,
		)
	} else if utype == "user" {
		// get all invitations sent by different owners to  specific user
		query = fmt.Sprintf("select owners.email, owners.channelname, invites.role from invites inner join owners on invites.sender = owners.id where invites.receiver = %d", id)
	}

	rows, err := pDB.PDB.QueryContext(context.Background(), query)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		invite := Invites{}
		if utype == "owner" {
			rows.Scan(&invite.Email, &invite.Role)
		} else if utype == "user" {
			rows.Scan(&invite.Email, &invite.ChannelName, &invite.Role)
		}
		list = append(list, invite)
	}
	return list, nil
}

type content struct {
	Id        int64
	update_at string
	// contentName
}

func (pDB *PostgreDB) AddContent(name string) error {
	query := fmt.Sprintf(`insert into contents(contentname) values('%s')`, name)
	_, err := pDB.PDB.ExecContext(context.Background(), query)
	if err != nil {
		return err
	}
	return nil
}

type OwnerProfile struct {
	ProfileImage string `json:"profileImage"`
	Title        string `json:"title"`
	ChannelUrl   string `json:"channelUrl"`
	Description  string `json:"description"`
}

func (pDB *PostgreDB) GetOwnerProfile(email string) (*OwnerProfile, error) {
	query := fmt.Sprintf("select channelname, channelurl, profileimage, description from owners where email = '%s'", email)
	row := pDB.PDB.QueryRowContext(context.Background(), query)
	u := &OwnerProfile{}

	err := row.Scan(
		&u.Title,
		&u.ChannelUrl,
		&u.ProfileImage,
		&u.Description,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, errors.New("user not found")
		default:
			return nil, errors.New("something went wrong")
		}
	}
	return u, nil
}

type teamDetailsLite struct {
	ChannelName  string `json:"channelName"`
	ProfileImage string `json:"profileImage"`
	OwnerId      string `json:"ownerId"`
}

func (pDB *PostgreDB) GetTeams(userId int64) (*[]teamDetailsLite, error) {
	query := fmt.Sprintf(
		"select owners.channelname, owners.profileimage, owners.id from teams inner join owners on teams.owner = owners.id where teams.member = %d",
		userId,
	)

	rows, err := pDB.PDB.QueryContext(context.Background(), query)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	teamDetLiteList := []teamDetailsLite{}
	for rows.Next() {
		teamDetLite := teamDetailsLite{}
		rows.Scan(&teamDetLite.ChannelName, &teamDetLite.ProfileImage, &teamDetLite.OwnerId)
		teamDetLiteList = append(teamDetLiteList, teamDetLite)
	}
	fmt.Println(teamDetLiteList)
	return &teamDetLiteList, nil
}

func (pDB *PostgreDB) CreateContent(name string, ownerId int64) error {
	query := fmt.Sprintf("insert into contents(projectname, owner) values('%s', %d)", name, ownerId)
	if _, err := pDB.PDB.ExecContext(context.Background(), query); err != nil {
		return err
	}
	return nil
}

type Content struct {
	Id int64 `json:"id"`
	Name string `json:"name"`
}

func (pDB *PostgreDB) GetContentList(ownerId int64) (*[]Content, error){
	contentList := new([]Content)
	query := fmt.Sprintf("select id, projectname from contents where owner = %d", ownerId)
	rows, err := pDB.PDB.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	for rows.Next(){
		content := new(Content)
		rows.Scan(&content.Id, &content.Name)
		*contentList = append(*contentList, *content)
	}
	return contentList, nil
} 

func (pDB *PostgreDB) GetEmailFromOwner(email string) (bool, int64, error) {
	query := fmt.Sprintf("select email,id from owners where email = '%s'", email)

	emailInDB := ""
	var idInDb int64
	row := pDB.PDB.QueryRowContext(context.Background(), query)
	err := row.Scan(&emailInDB, &idInDb)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return false, -1, nil
		default:
			return false, -1, err
		}
	}
	return true,idInDb, nil
}

func (pDB *PostgreDB) IsTeamMember(ownerId int64, userId int64) (bool, error){
	query := fmt.Sprintf("select id from teams where owner = %d AND member = %d", ownerId, userId);
	row := pDB.PDB.QueryRowContext(context.Background(), query)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		switch err {
		case sql.ErrNoRows: return false, nil
		default: return false, err
		}
	}
	return true, nil
}

type contentDetail struct {
	ProjectName string `json:"projectName"`
	Title string `json:"title"`
	Description string `json:"description"`
	Tags string `json:"tags"`
	Video string `json:"video"`
}

func (pDB *PostgreDB) GetContentDetail(contentId int64) (*contentDetail, error){
	contentDetail := new(contentDetail)
	query := fmt.Sprintf("select projectname, title, video, description, tags from contents where id = %d", contentId)
	row := pDB.PDB.QueryRowContext(context.Background(), query)
	err := row.Scan(&contentDetail.ProjectName, &contentDetail.Title, &contentDetail.Video, &contentDetail.Description, &contentDetail.Tags)
	if err != nil {
		switch err {
		case sql.ErrNoRows: return nil, errors.New("content not found")
		default: return nil, err
		}
	}
	return contentDetail, nil
}
