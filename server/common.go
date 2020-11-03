package server

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gmo-personal/picshare_pic_service/database"
	"github.com/gmo-personal/picshare_pic_service/model"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Headers", "Authorization")
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT")
}

func InitServer() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/upload/", uploadHandler)
	http.HandleFunc("/list/", listHandler)
	http.HandleFunc("/listUser/", listHandlerUser)
	http.HandleFunc("/count/", countHandler)
	http.HandleFunc("/display/", displayHandler)
	http.HandleFunc("/delete/", deleteHandler)
	http.HandleFunc("/hide/", hideHandler)
	http.HandleFunc("/like/", likeHandler)
	http.HandleFunc("/likesCount/", likesCountHandler)
	http.HandleFunc("/likesInfo/", likesInfoHandler)
	http.HandleFunc("/date/", dateHandler)
	http.HandleFunc("/hideStatus/", hideStatusHandler)
	http.HandleFunc("/favorite/", favoriteHandler)
	http.HandleFunc("/writeFav/", writeIsFavoritedHandler)
	http.HandleFunc("/writeTotalFav/", writeFavoritesHandler)
	http.HandleFunc("/grey/", greyHandler)


	log.Fatal(http.ListenAndServe(":8082", nil))
}

func writeIsFavoritedHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	link := getURLParam(r,"src")
	picId := database.GetPicIdFromLink(link)
	userId := GetUserIdFromRequest(w,r)
	if userId == -1 {
		fmt.Println("Error in UserId Request Write is Favorite Handler")
	}

	isFavorited := database.GetIsFavorited(picId, userId)
	fmt.Fprintln(w, isFavorited +" ")
}

func writeFavoritesHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	userId := GetUserIdFromRequest(w,r)
	fmt.Println(userId)
	if userId == -1 {
		fmt.Println("Error in UserId Request Write Favorites Handler")
		return
	}
	Favorites := database.GetTotalFavorited(userId)
	fmt.Println(Favorites)
	fmt.Fprintln(w, Favorites +" ")
}

func hideStatusHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	link := getURLParam(r,"src")
	picId := database.GetPicIdFromLink(link)
	hideStatus := database.CheckHidden(picId)
	if hideStatus == - 1 {
		fmt.Println("Error in HideStatus")
		return
	} else {
		w.WriteHeader(http.StatusOK)
	}

	fmt.Fprintln(w, strconv.Itoa(hideStatus))
}

func countHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	userId := GetUserIdFromRequest(w, r)
	if userId == -1 {
		return
	}
	count := database.CountUserPics(userId)
	fmt.Fprintln(w, count)
}

func dateHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	link := getURLParam(r,"src")
	picId := database.GetPicIdFromLink(link)
	date := database.GetDateFromPicId(picId)
	if date == "" {
		fmt.Println("Error in Date")
		return
	} else {
		w.WriteHeader(http.StatusOK)
	}

	fmt.Fprintln(w, date)
}

func likesCountHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	link := getURLParam(r,"src")
	picId := database.GetPicIdFromLink(link)
	count := database.CountPicLikes(picId)
	if count == -1 {
		fmt.Println("Error in Count")
		return
	} else {
		w.WriteHeader(http.StatusOK)
	}
	userId := GetUserIdFromRequest(w,r)

	isLiked := database.GetIsLiked(picId, userId)
	fmt.Fprintln(w, isLiked +" "+ strconv.Itoa(count))
}

func likesInfoHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	userId := GetUserIdFromRequest(w,r)
	if userId == -1 {
		return
	}

	likesReceived := database.GetLikesReceived(userId)
	if likesReceived == - 1 {
		fmt.Println("Error retrieving likes received")
		return
	}

	likesSent := database.GetLikesSent(userId)
	if likesSent == - 1 {
		fmt.Println("Error retrieving likes Sent")
		return
	}
	fmt.Fprintln(w, strconv.Itoa(likesSent) + " " + strconv.Itoa(likesReceived))
}

func displayHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	link := getURLParam(r,"src")
	returnLink := database.GetFullLink(link)
	if len(returnLink) == 0 {
		w.WriteHeader(http.StatusNotFound)
	}
	fmt.Fprintln(w, returnLink)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	userId := GetUserIdFromRequest(w, r)
	if userId == -1 {
		return
	}
	link := getURLParam(r,"src")
	userExpectedId := database.GetUserIdFromLink(link)

	if userId != userExpectedId {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := database.SetPicDeleted(link)
	if err == "" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusPreconditionFailed)
	}
}

func likeHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	userId := GetUserIdFromRequest(w, r)

	if userId == -1 {
		return
	}

	link := getURLParam(r,"src")

	res := database.LikePic(userId, link)
	if  res != "true" && res !="false" {
		w.WriteHeader(http.StatusPreconditionFailed)
	}

	picId := database.GetPicIdFromLink(link)
	count := database.CountPicLikes(picId)
	if count == -1 {
		fmt.Println("Error in Count")
		return
	} else {
		w.WriteHeader(http.StatusOK)
	}

	fmt.Fprintln(w, res + " " + strconv.Itoa(count))
}

func favoriteHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	userId := GetUserIdFromRequest(w, r)

	if userId == -1 {
		return
	}

	link := getURLParam(r,"src")

	res := database.FavoritePic(userId, link)
	if  res != "true" && res !="false" {
		w.WriteHeader(http.StatusPreconditionFailed)
	}

	fmt.Fprintln(w, res + " ")
}

func hideHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	userId := GetUserIdFromRequest(w, r)
	if userId == -1 {
		//crashes if unauthorized goes here
		return
	}
	link := getURLParam(r,"src")
	userExpectedId := database.GetUserIdFromLink(link)

	if userId != userExpectedId {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	picId := database.GetPicIdFromLink(link)
	if picId == -1 {
		fmt.Println("Failed to retrieve ID")
		return
	}
	hide := database.HideImage(picId)
	if hide != "" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusPreconditionFailed)
	}

	fmt.Fprintf(w, hide)
}

func getURLParam(r *http.Request, paramName string) string {
	keys, seen := r.URL.Query()[paramName]

	if seen && len(keys) > 0 {
		return keys[0]
	}
	return ""
}

func getURLIntParam(r *http.Request, paramName string) int {
	paramStr := getURLParam(r, paramName)

	paramInt, err := strconv.Atoi(paramStr)
	if err != nil {
		return -1
	}
	return paramInt
}

const (
	AWS_S3_REGION = "us-east-2"
	AWS_S3_BUCKET_BASE = "pic-share-"
)

var sess = connectAWS()
var uploader = s3manager.NewUploader(sess)
var publicRead = "public-read"

func connectAWS() *session.Session {
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(AWS_S3_REGION),
		},
	)
	if err != nil {
		panic(err)
	}
	return sess
}

func greyHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	src := getURLParam(r,"src")
	rM, _ := strconv.ParseFloat(getURLParam(r,"r"), 64)
	bM, _  := strconv.ParseFloat(getURLParam(r,"b"), 64)
	gM, _  := strconv.ParseFloat(getURLParam(r,"g"), 64)
	fmt.Println(src, rM, bM, gM)

	picToGreyscale, err := http.Get(src)
	if err != nil {
		fmt.Println("Error Retrieving the picture")
		fmt.Println(err)
		return
	}
	defer picToGreyscale.Body.Close()

	img, format, err := image.Decode(picToGreyscale.Body)
	if err != nil {
		fmt.Println("greyHandler:  " , err)
		return
	}

	greyScale := greyScale(img, rM, bM, gM)

	file := new(bytes.Buffer)
	if format == "jpeg" {
		err = jpeg.Encode(file, greyScale, nil)
		fmt.Println("Jpg Encode Success!")
	} else if format =="png" {
		err = png.Encode(file, greyScale)
		fmt.Println("PNG Encode Success!")
	} else {
		fmt.Println("greyHandler: " , format)
		return
	}
	if err!= nil {
		fmt.Println(err)
	}

	reader := bufio.NewReader(file)
	content, _ := ioutil.ReadAll(reader)
	encoded := base64.StdEncoding.EncodeToString(content)
	w.Write([]byte(encoded))

}

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)
	userId := GetUserIdFromRequest(w, r)

	_ = r.ParseMultipartForm(10 << 23)
	pic, _, err := r.FormFile("pic")
	if err != nil {
		fmt.Println("Error Retrieving the picture")
		fmt.Println(err)
		return
	}
	defer pic.Close()

	imageId := database.InsertPic(&model.Image{
		UserId:    userId,
		Deleted:   false,
		CreatedAt: "",
	})

	picToCrop, _, err := r.FormFile("pic")
	if err != nil {
		fmt.Println("Error Retrieving the picture")
		fmt.Println(err)
		return
	}
	defer picToCrop.Close()


	picToGreyscale1, _, err := r.FormFile("pic")
	if err != nil {
		fmt.Println("Error Retrieving the picture")
		fmt.Println(err)
		return
	}
	defer picToGreyscale1.Close()


	passed := uploadBucket("original", imageId, pic)
	if passed != true {
		fmt.Println("Error uploading bucket original")
		w.WriteHeader(http.StatusConflict)
		return
	}


	cropped := cropImage(picToCrop)
 	passed = uploadBucket("crop", imageId, cropped)
	if passed != true {
		fmt.Println("Error uploading bucket crop")
		w.WriteHeader(http.StatusConflict)
		return
	}

	greyScale1 := greyScaleImage(picToGreyscale1,0.92126, 0.97152,0.90722)

	passed = uploadBucket("greyscale-1", imageId, greyScale1)
	if passed != true {
		fmt.Println("Error uploading bucket greyScale")
		w.WriteHeader(http.StatusConflict)
		return
	}

	email := database.RetrieveEmail(imageId)
	if len(email) != 0 {
		message := []byte("To:" + email +"\r\n" +
			"Subject: Upload to PicShare successful!\r\n" +
			"\r\n" +
			"Your image has been successfully uploaded!\r\n\n"  )
		fmt.Println(message)
		////// TURNED OFF FOR NOW
		//SendEmail(email, message)
		//////
	}
	w.WriteHeader(http.StatusOK)
}

func uploadBucket(format string, imageId int, pic io.Reader) bool {
	fmt.Println("enteredUploading")
	output, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(AWS_S3_BUCKET_BASE + format), // Bucket to be used
		Key:    aws.String(strconv.Itoa(imageId)+ format +".jpg"),      // Name of the file to be saved
		Body:   pic,                      // File
		//ContentType: aws.String("image/jpeg"),
		ACL: &publicRead,
	})

	if err != nil {
		fmt.Println(err)// Do your error handling here
		return false
	}
	link := output.Location
	passed := database.UploadLink(imageId, link, format)
	if format == "crop" {
		resizedLink := strings.Split(link, "-crop")[0] + "-crop-resized" + strings.Split(link, "-crop")[1]
		passed = database.UploadLink(imageId, resizedLink, "resized")
	}
	return passed
}

func cropImage (pic io.Reader) io.Reader {

	img, format, err := image.Decode(pic)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	cropImage := cropCenter(img)

	file := new(bytes.Buffer)
	if format == "jpeg" {
		jpeg.Encode(file, cropImage, nil)
	} else if format =="png" {
		png.Encode(file, cropImage)
	} else {
		fmt.Println(format)
		return nil
	}

	return  file
}

func cropCenter(img image.Image) image.Image {
	rect := img.Bounds()
	dimensions := min(rect.Dx(),rect.Dy())
	top := (rect.Dy() - dimensions) / 2
	left := (rect.Dx() - dimensions) / 2

	cropImage := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(left, top, left + dimensions, top + dimensions))

	return cropImage
}

func greyScaleImage (pic io.Reader, rM, bM, gM float64) io.Reader {

	img, format, err := image.Decode(pic)
	if err != nil {
		fmt.Println("line 422 -> " , err)
		return nil
	}

	greyScale := greyScale(img, rM, bM, gM)

	file := new(bytes.Buffer)
	if format == "jpeg" {
		jpeg.Encode(file, greyScale, nil)
	} else if format =="png" {
		png.Encode(file, greyScale)
	} else {
		fmt.Println("line 434->" , format)
		return nil
	}
	return  file
}

func greyScale(img image.Image, rM, bM, gM float64) image.Image {
	size := img.Bounds().Size()
	rect := image.Rect(0, 0, size.X, size.Y)
	wImg := image.NewRGBA(rect)


	for x := 0; x < size.X; x++ {
		// and now loop thorough all of this x's y
		for y := 0; y < size.Y; y++ {
			pixel := img.At(x, y)
			originalColor := color.RGBAModel.Convert(pixel).
			(color.RGBA)
			// Offset colors a little, adjust it to your taste
			r := float64(originalColor.R) * rM
			g := float64(originalColor.G) * gM
			b := float64(originalColor.B) * bM
			// average
			grey := uint8((r + g + b) / 3)
			c := color.RGBA{
				R: grey, G: grey, B: grey, A: originalColor.A,
			}

			wImg.Set(x, y, c)
		}
	}

	return wImg
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	sLikes := getURLParam(r, "sLikes")
	sFaves := getURLParam(r, "sFaves")

	userId := GetUserIdFromRequest(w, r)
	if userId == -1 {
		sFaves = "false"
	}

	var linkList = database.ListGlobalPics(userId, sLikes, sFaves)
	fmt.Fprintln(w, linkList)
}

var client = http.Client{}

func listHandlerUser(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	sLikes := getURLParam(r, "sLikes")
	sFaves := getURLParam(r, "sFaves")
	userId := GetUserIdFromRequest(w, r)
	if userId == -1 {
		return
	}

	var linkList = database.ListUserPics(userId, sLikes, sFaves)
	fmt.Fprintln(w, linkList)
}

func GetUserIdFromRequest(w http.ResponseWriter, r *http.Request ) int{
	enableCors(&w)
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) == 1 {
		fmt.Println("Token is Empty")
		return -1
	}
	reqToken = splitToken[1]

	req, err := http.NewRequest("GET", "http://localhost:8081/validate/", nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Authorization", "Bearer " + reqToken)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return -1
	}
	userIdString := string(bodyBytes)

	userId, err := strconv.Atoi(userIdString)

	if err != nil {
		fmt.Println(err)
		return -1
	}
	return userId
}

func SendEmail(receiver string, message []byte) {
	from := "resforge.dev@gmail.com"
	password := "!Andy***951"

	to := []string{
		receiver,
	}

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Email Sent Successfully!")
}
