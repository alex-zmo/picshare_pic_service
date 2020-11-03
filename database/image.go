package database

import (
	"database/sql"
	"fmt"
	"github.com/gmo-personal/picshare_pic_service/model"
	"strconv"
)

func createImageTable() {
	createImageStmt := `CREATE TABLE IF NOT EXISTS pic (
		id INT AUTO_INCREMENT,
		user_id INT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted TINYINT(1) DEFAULT 0,
		hidden TINYINT(1) DEFAULT 0,
		likes INT DEFAULT 0,
		PRIMARY KEY (id),
		FOREIGN KEY (user_id) REFERENCES user(id)
	);`

	_, err := db.Exec(createImageStmt)
	if err != nil {
		fmt.Println(err)
	}
}


func createLinkTable() {
	createLinkStmt := `CREATE TABLE IF NOT EXISTS link (
		id INT AUTO_INCREMENT,
		pic_id INT,
		link VARCHAR(256),
		format varchar(32),
		PRIMARY KEY (id),
		FOREIGN KEY (pic_id) REFERENCES pic(id)
	);`

	_, err := db.Exec(createLinkStmt)
	if err != nil {
		fmt.Println(err)
	}
}

func createLikesTable() {
	createLikesStmt := `CREATE TABLE IF NOT EXISTS likes (
		user_id INT,
		pic_id INT,
		liked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, pic_id),
		FOREIGN KEY (user_id) REFERENCES user(id),
		FOREIGN KEY (pic_id) REFERENCES pic(id)
	);`
	_, err := db.Exec(createLikesStmt)
	if err != nil {
		fmt.Println(err)
	}
}

func createFavoritesTable() {
	createFavoritesStmt := `CREATE TABLE IF NOT EXISTS favorites (
		user_id INT,
		pic_id INT,
		favorited_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, pic_id),
		FOREIGN KEY (user_id) REFERENCES user(id),
		FOREIGN KEY (pic_id) REFERENCES pic(id)
	);`
	_, err := db.Exec(createFavoritesStmt)
	if err != nil {
		fmt.Println(err)
	}
}


func FavoritePic(userId int, src string) string {
	checkFavoriteTableStmt :=`SELECT COUNT(pic_id) FROM favorites WHERE pic_id = ? AND user_id = ?`

	picId := GetPicIdFromLink(src)
	if picId == - 1 {
		return "Error: Pic ID Not Valid"
	}

	count, err := db.Query(checkFavoriteTableStmt, picId, userId)
	if err != nil {
		fmt.Println(err)
		return "Error Count Pic Error"
	}
	defer closeRows(count)

	exists := -1
	if count.Next() {
		err := count.Scan(&exists)
		if err != nil {
			fmt.Println(err)
		}
	}

	result := ""
	if exists == 1 {
		deleteFavorite(userId, picId)
		result = "false"
	} else if exists == 0 {
		addFavorite(userId, picId)
		result = "true"
	}

	return result
}

func deleteFavorite(userId, picId int) {
	DeleteFavoriteStmt :=`DELETE FROM favorites WHERE user_id = ? AND pic_id = ?`
	_, err := db.Exec(DeleteFavoriteStmt, userId, picId)
	if err != nil {
		fmt.Println(err)
	}
}

func addFavorite(userId, picId int) {
	AddFavoriteTableStmt :=`INSERT INTO favorites(user_id, pic_id) VALUES(?, ?)`
	_, err := db.Exec(AddFavoriteTableStmt, userId, picId)
	if err != nil {
		fmt.Println(err)
	}
}

func InsertPic(picInfo *model.Image) int {
	insertImageStmt := `INSERT INTO pic (		
		user_id
	) VALUES (?);`

	result, err := db.Exec(insertImageStmt, picInfo.UserId)
	if err != nil {
		fmt.Println(err)
	}
	id, _ := result.LastInsertId()
	return int(id)
}

func ListGlobalPics(userId int, sLikes, sFaves string) []string {

	listAllImageStmt := `SELECT l.link
	FROM pic p LEFT JOIN link l
	ON p.id = l.pic_id
	WHERE p.deleted = 0 
	AND p.hidden = 0
	AND l.format = 'resized'
`
	listAllImageWithFavesStmt := `
	SELECT l.link
	FROM pic p LEFT JOIN link l
	ON p.id = l.pic_id
	WHERE p.deleted = 0
	AND p.hidden = 0
	AND l.format = 'resized'
	AND pic_id IN (
	SELECT pic_id FROM favorites WHERE user_id = ?)`

	if sLikes == "true" {
		listAllImageStmt += `ORDER BY likes DESC`
		listAllImageWithFavesStmt += `ORDER BY likes DESC`
	}


	result := make([]string,0)
	var res *sql.Rows
	var err error
	if sFaves == "true" {
		res, err = db.Query(listAllImageWithFavesStmt, userId)
	} else {
		res, err = db.Query(listAllImageStmt)
	}
	if err != nil {
		fmt.Println(err)
		return result
	}
	defer closeRows(res)

	link := ""
	for res.Next() {
		res.Scan(&link)
		result = append(result, link)
	}
	return result
}

func ListUserPics(userId int, sLikes, sFaves string) []string {

	listImageStmt := `SELECT l.link
	FROM pic p LEFT JOIN link l
	ON p.id = l.pic_id
	WHERE p.deleted = 0
	AND p.user_id = ?
	AND l.format = 'resized'
	`

	listImageWithFavesStmt := `
	SELECT l.link
	FROM pic p LEFT JOIN link l
	ON p.id = l.pic_id
	WHERE p.deleted = 0
	AND p.user_id = ?
	AND l.format = 'resized'
	AND pic_id IN (
	SELECT pic_id FROM favorites WHERE user_id = ?)`


	if sLikes == "true" {
		listImageStmt += `ORDER BY likes DESC`
		listImageWithFavesStmt += `ORDER BY likes DESC`
	}

	result := make([]string,0)
	var res *sql.Rows
	var err error
	if sFaves == "true" {
		res, err = db.Query(listImageWithFavesStmt, userId, userId)
	} else {
		res, err = db.Query(listImageStmt, userId)
	}
	defer closeRows(res)

	link := ""
	for res.Next() {
		res.Scan(&link)
		result = append(result,link)
	}
	fmt.Println(err)
	return result
}

func CountUserPics(userId int) int {

	listImageStmt := `SELECT COUNT(l.link)
	FROM pic p LEFT JOIN link l
	ON p.id = l.pic_id
	WHERE p.deleted = 0
	AND p.user_id = ?
	AND l.format = 'crop'
	`

	res, err := db.Query(listImageStmt, userId)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	defer closeRows(res)
	count := 0
	if res.Next() {
		res.Scan(&count)
	}
	return count
}

func UploadLink(imageId int, link, format string) bool {
	uploadLinkStmt := `INSERT INTO link (pic_id, link, format) values (?, ?, ?)`
	_, err := db.Exec(uploadLinkStmt, imageId, link, format)
	if err != nil {
		fmt.Println("Error in Uploading")
		return false
	}
	return true
}

func RetrieveEmail(imageId int) string {
	retrieveEmailStmt := `SELECT email FROM user
	WHERE id = (
	SELECT user_id 
	from link l
	INNER JOIN pic p
	ON l.pic_id = p.id
	AND l.pic_id = ?
	LIMIT 1
	)
	`
	res, err := db.Query(retrieveEmailStmt, imageId)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer closeRows(res)
	email := ""
	if res.Next() {
		err := res.Scan(&email)
		if err != nil {
			fmt.Println(err)
		}
	}
	return email
}

func GetFullLink(link string) string {
	retrieveLinkStmt := `SELECT link FROM link
	WHERE format = 'original' 
	AND pic_id = (SELECT pic_id FROM link WHERE link = ? LIMIT 1)
	`
	res, err := db.Query(retrieveLinkStmt, link)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer closeRows(res)
	linkResult := ""
	if res.Next() {
		err := res.Scan(&linkResult)
		if err != nil {
			fmt.Println(err)
		}
	}
	return linkResult
}


func SetPicDeleted(link string) string {
	setPicDeletedStmt := `UPDATE pic SET deleted = 1 WHERE id= ?`

	picId := GetPicIdFromLink(link)
	if picId == -1 {
		fmt.Println("Failed to retrieve ID")
		return "Failed to retrieve ID"
	}
	_, err := db.Exec(setPicDeletedStmt, picId)
	if err != nil {
		fmt.Println(err)
		return "Delete Failed"
	}
	return ""
}

func GetPicIdFromLink(link string) int {
	retrieveLinkStmt := `SELECT pic_id FROM link WHERE link = ? LIMIT 1
	`
	res, err := db.Query(retrieveLinkStmt, link)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer closeRows(res)

	picId := -1
	if res.Next() {
		err := res.Scan(&picId)
		if err != nil {
			fmt.Println(err)
		}
	}
	return picId
}

func HideImage(picId int) string{
	hidden := CheckHidden(picId)
	hide := -1
	if hidden == 1 {
		hide = 0
	} else if hidden == 0 {
		hide = 1
	} else {
		return ""
	}

	hidePicStmt :=`UPDATE pic SET hidden = ? WHERE id = ?`
	_, err := db.Exec(hidePicStmt, hide , picId )
	if err != nil {
		fmt.Println(err)
		return "Hide Failed"
	}
	return strconv.Itoa(hide)
}

func CheckHidden(picId int) int {
	checkHiddenStmt := `SELECT hidden FROM pic WHERE id = ?`
	res, err := db.Query(checkHiddenStmt, picId)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer closeRows(res)

	hidden := -1
	if res.Next() {
		err := res.Scan(&hidden)
		if err != nil {
			fmt.Println(err)
		}
	}
	return hidden
}

func GetUserIdFromLink(link string) int {
	retrieveUserIdStmt := `SELECT user_id FROM pic 
	WHERE id = (SELECT pic_id 
	FROM link WHERE link = ? LIMIT 1)
	`

	res, err := db.Query(retrieveUserIdStmt, link)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer closeRows(res)

	userId := -1
	if res.Next() {
		err := res.Scan(&userId)
		if err != nil {
			fmt.Println(err)
		}
	}
	return userId
}

func LikePic(userId int, src string) string {
	checkLikesTableStmt :=`SELECT COUNT(pic_id) FROM likes WHERE pic_id = ? AND user_id = ?`

	picId := GetPicIdFromLink(src)
	if picId == - 1 {
		return "Error: Pic ID Not Valid"
	}

	count, err := db.Query(checkLikesTableStmt, picId, userId)
	if err != nil {
		fmt.Println(err)
		return "Error Count Pic Error"
	}
	defer closeRows(count)

	exists := -1
	if count.Next() {
		err := count.Scan(&exists)
		if err != nil {
			fmt.Println(err)
		}
	}

	result := ""
	if exists == 1 {
		deleteLike(userId, picId)
		result = "false"
	} else if exists == 0 {
		addLike(userId, picId)
		result = "true"
	}

	UpdateTotalLikes(picId)
	return result
}

func deleteLike(userId, picId int) {
	DeleteLikesTableStmt :=`DELETE FROM likes WHERE user_id = ? AND pic_id = ?`
	_, err := db.Exec(DeleteLikesTableStmt, userId, picId)
	if err != nil {
		fmt.Println(err)
	}
}

func addLike(userId, picId int) {
	AddLikesTableStmt :=`INSERT INTO likes(user_id, pic_id) VALUES(?, ?)`
	_, err := db.Exec(AddLikesTableStmt, userId, picId)
	if err != nil {
		fmt.Println(err)
	}
}

func UpdateTotalLikes(picId int) {
	addTotalLikesStmt :=`UPDATE pic SET likes =  
	(SELECT COUNT(user_id) 
	FROM likes WHERE pic_id = ?) WHERE id = ?`
	_, err := db.Exec(addTotalLikesStmt, picId, picId)
	if err != nil {
		fmt.Println(err)
	}
}

func CountPicLikes(picId int) int {
	getLikesStmt := `SELECT likes FROM pic where id = ?`
	count, err := db.Query(getLikesStmt, picId)
	if err != nil {
		fmt.Println(err)
		return - 1
	}
	defer closeRows(count)

	likes := 0
	if count.Next() {
		err := count.Scan(&likes)
		if err != nil {
			fmt.Println(err)
		}
	}
	return likes
}

func GetIsLiked(picId, userId int) string {
	getIsLikedStmt := `SELECT COUNT(user_id) FROM likes WHERE pic_id = ? AND user_id = ? LIMIT 1`
	count, err := db.Query(getIsLikedStmt, picId, userId)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer closeRows(count)

	liked := 0
	if count.Next() {
		err := count.Scan(&liked)
		if err != nil {
			fmt.Println(err)
		}
	}

	result := ""
	if liked == 1 {
		return "true"
	} else if liked == 0 {
		return "false"
	}

	return result
}

func GetTotalFavorited(userId int) string {
	getTotalFavoritedStmt := `SELECT COUNT(pic_id) FROM favorites WHERE user_id = ? LIMIT 1`
	count, err := db.Query(getTotalFavoritedStmt, userId)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer closeRows(count)

	fav := 0
	if count.Next() {
		err := count.Scan(&fav)
		if err != nil {
			fmt.Println(err)
		}
	}
	return strconv.Itoa(fav)
}

func GetIsFavorited(picId, userId int) string {
	getIsFavoritedStmt := `SELECT COUNT(user_id) FROM favorites WHERE pic_id = ? AND user_id = ? LIMIT 1`
	count, err := db.Query(getIsFavoritedStmt, picId, userId)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer closeRows(count)

	fav := 0
	if count.Next() {
		err := count.Scan(&fav)
		if err != nil {
			fmt.Println(err)
		}
	}

	result := ""
	if fav == 1 {
		return "true"
	} else if fav == 0 {
		return "false"
	}

	return result
}

func GetLikesReceived(userId int) int {
	getLikesReceivedStmt := `SELECT SUM(likes) FROM pic where user_id = ?`
	count, err := db.Query(getLikesReceivedStmt,  userId)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer closeRows(count)

	liked := 0
	if count.Next() {
		err := count.Scan(&liked)
		if err != nil {
			fmt.Println(err)
		}
	}
	return liked
}


func GetLikesSent(userId int) int {
	getLikesSentStmt := `SELECT COUNT(pic_id) FROM likes where user_id = ?`
	count, err := db.Query(getLikesSentStmt,  userId)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer closeRows(count)

	liked := 0
	if count.Next() {
		err := count.Scan(&liked)
		if err != nil {
			fmt.Println(err)
		}
	}
	return liked
}

func GetDateFromPicId(picId int) string {
	getDateStmt := `SELECT created_at FROM pic WHERE id = ?`
	count, err := db.Query(getDateStmt,  picId)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer closeRows(count)

	liked := ""
	if count.Next() {
		err := count.Scan(&liked)
		if err != nil {
			fmt.Println(err)
		}
	}
	return liked
}