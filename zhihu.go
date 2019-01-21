package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"

	//"sync"
	"bufio"
	"strconv"
	"time"
	//"runtime"
)

var quit chan int = make(chan int)

func main() {

	url := urlExists()

	dirName := strconv.FormatInt(time.Now().Unix(), 10)
	imgDir := "./" + dirName + "/"

	_, errs := pathExists(imgDir)
	if errs != nil {
		fmt.Printf("创建文件失败")
		return
	}

	getImgURL(url, imgDir)
}

//获取图片地址
func getImgURL(url, imgDir string) {
	//提交请求
	reqest, err := http.NewRequest("GET", url, nil)

	//增加header选项
	reqest.Header.Add("Content-Type", "application/json; charset=utf-8")
	//reqest.Header.Add("Cookie", "_zap=97a733fe-7d3e-45e0-9fe8-c138f01b6fc9; d_c0=\"AEBjiMYYoA2PTo-4jA6iymyc8bkORSrQyVg=|1526862212\"; _xsrf=rSY6NLBWiJVpd3t779cxSpkzOCTNZfnn; z_c0=Mi4xTURUVEFRQUFBQUFBUUdPSXhoaWdEUmNBQUFCaEFsVk4xakhKWEFBcmtpS09jQUFoVWpJM2NXNG9LQjBtbWxuWm5B|1541137366|ad106a49adcbac59f43323b0d4f97ded81c4a08b; __utmv=51854390.100-1|2=registration_date=20150706=1^3=entry_date=20150706=1; _ga=GA1.2.632016254.1541137374; tst=r; q_c1=c7a254833a0146769ea4f47f755774ba|1547085195000|1527140571000; __utmc=51854390; __utma=51854390.632016254.1541137374.1547085197.1547112488.3; __utmz=51854390.1547112488.3.3.utmcsr=zhihu.com|utmccn=(referral)|utmcmd=referral|utmcct=/people/zhao-gui-bin-zero/following/collections; tgw_l7_route=66cb16bc7f45da64562a077714739c11; arp_scroll_position=192745")
	reqest.Header.Add("Host", "www.zhihu.com")
	reqest.Header.Add("Referer", "www.zhihu.com")
	reqest.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36")

	if err != nil {
		panic(err)
	}

	//生成client 参数为默认
	client := &http.Client{}

	//处理返回结果
	response, err := client.Do(reqest)

	if err != nil {
		fmt.Println("图片地址不正确，请重试")
		return
	}

	defer response.Body.Close()

	if response.StatusCode == 200 {
		body, _ := ioutil.ReadAll(response.Body)
		bodystr := string(body)

		//下一页数据
		next := gjson.Get(bodystr, "paging.next")

		fmt.Println(next.String() + "------------------------------------------------------")

		for i := 0; i < 5; i++ {
			value := gjson.Get(bodystr, "data."+strconv.Itoa(i)+".content")

			if len(value.String()) > 0 {
				var re = regexp.MustCompile(`(?m)<noscript><img src="(.\S*)"`)

				for _, match := range re.FindAllStringSubmatch(value.String(), -1) {
					go downloadImg(match[1], imgDir)
					<-quit
				}

			} else {
				fmt.Println("抓取结束")
				return
			}
		}

		if len(next.String()) > 0 {
			getImgURL(next.String(), imgDir)
		}else{
			fmt.Println("抓取结束，已到最后一页")
			return
		}


	} else {
		fmt.Println("获取数据错误，请重试")
		return
	}
}

//下载图片
func downloadImg(url, imgDir string) (n int64, err error) {
	path := strings.Split(url, "/")
	var name string
	if len(path) > 1 {
		name = path[len(path)-1]
	}

	fmt.Println(name)
	out, err := os.Create(imgDir + name)
	defer out.Close()
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println("图片下载失败")
		quit <- 0
		return
	}

	quit <- 0

	defer resp.Body.Close()
	pix, err := ioutil.ReadAll(resp.Body)
	n, err = io.Copy(out, bytes.NewReader(pix))

	return
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else {
		err := os.Mkdir(path, os.ModePerm)

		if err != nil {
			return false, nil
		}
	}

	return true, nil
}

//判断URL是否为空
func urlExists() string {
	urls := bufio.NewReader(os.Stdin)
	fmt.Print("输入Url:")
	url, _, _ := urls.ReadLine()
	urlStr := string(url[:])

	if len(url) == 0 {
		urlStr = urlExists()
	}

	return urlStr

}
