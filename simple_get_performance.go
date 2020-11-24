package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

type urlInfo struct {
	URL       string `json:"URL"` // Json 데이터와 annotation mapping
	APIKey    string `json:"API_KEY"`
	SpeakerID string `json:"Speaker_id"`
}

type receiveStruct struct {
	Datatype string    `json:"type"` // Json 데이터와 annotation mapping
	Wav      []float64 `json:"wav"`
	Text     string
}

// build unique filename
func buildFileName() string {
	now := time.Now()
	return now.Format("2006-01-02 15:04:05")
	// return now.Format(time.RFC3339)
}

func reqWave(s string, wg *sync.WaitGroup, conf *urlInfo) {
	defer wg.Done()

	fmt.Println("Running Sentences = ", s)
	fmt.Println("Running Key = ", conf.APIKey)
	req, err := http.NewRequest("GET", conf.URL, nil)
	if err != nil {
		log.Print("Error is = ", err)
		os.Exit(1)
	}

	q := req.URL.Query()
	q.Add("api_key", conf.APIKey)
	q.Add("speaker_id", conf.SpeakerID)
	q.Add("sentence", s)

	req.URL.RawQuery = q.Encode()
	fmt.Println(req.URL.String())

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			fmt.Println("redirectPolicyFunc()")
			return http.ErrUseLastResponse // 자동 리다이렉트 하지 않음
		},
	}

	// request start
	start := time.Now()

	resp, _ := client.Do(req)

	elapsed := time.Since(start)
	log.Printf("Execution time %s %d ", elapsed, elapsed)
	// request end

	// Check Response
	fmt.Println(resp)
	// fmt.Println(resp.Body)

	bytes, _ := ioutil.ReadAll(resp.Body)

	// Case1. 출력을 위해 byte -> string 처리
	// data := string(bytes)
	// fmt.Println(data)

	// Case2. byte -> struct (Unmarshalling)
	// var rcv receiveStruct
	// rcv := receiveStruct{}
	rcv := new(receiveStruct)
	rcv.Text = s
	json.Unmarshal(bytes, &rcv)
	// fmt.Printf("%+v\n", rcv)
	// fmt.Printf("%+v\n", sc)
	fmt.Printf("length = %d\n", len(rcv.Wav))

	// Case3. Bytes -> write File
	fn := "rcv/" + buildFileName() + ".json"
	err = ioutil.WriteFile(fn, bytes, 0644)
	if err != nil {
		log.Print("Error is = ", err)
		os.Exit(1)
	}

	// Case4. Json data -> Marshalling (struct -> byte) -> Save
	marshalbytes, _ := json.Marshal(rcv)
	fn2 := "rcv/" + buildFileName() + "_marshal_" + ".json"
	err = ioutil.WriteFile(fn2, marshalbytes, 0644)
	if err != nil {
		log.Print("Error is = ", err)
		os.Exit(1)
	}

	// 요청 Body 닫기
	defer resp.Body.Close()
}

func main() {
	runtime.GOMAXPROCS(8)
	fmt.Println(runtime.NumCPU()) // 8
	var wg sync.WaitGroup

	// sentenceList := []string{"안녕하세요 처음뵙겠습니다",
	// 	"안녕하십니까? 생명보험 상담원 김나래입니다.",
	// 	"이일범 고객님, 통화 가능하세요?",
	// 	"잠깐 통화 괜찮으세요? 소중한 시간 내주셔서 감사드립니다.",
	// 	"고객님 아시다시피 우리나라 부동의 사망률 일 위를 지키고 있는 암에 대해서 진단금이 부족한 경우가 굉장히 많으셨어요",
	// 	"연령이 높아지시면서 약이 한 두 개 추가가 되고, 그때 되어서 걱정된다 가입해야지 하시면 가입이 안되시거든요.",
	// 	"고객님의 정보는 저장 되어 있으시고, 고객님 본인이 맞는지 확인하기 위하여 이로 시작하는 주민번호 뒤의 일곱자리 말씀해주시겠어요?",
	// 	"이 보험은 만기지급금이 없는 순수보장성 보험입니다",
	// 	"상품의 보장 내용과 갱신 등에 관한 자세한 사항은 약관을 꼭 확인하시기 바랍니다",
	// 	"현재까지 치료받은 내용과 향후 치료계획이 있으신 경우 정확히 알려주세요",
	// 	"보험의 중요사항 및 면책사항은 보내드리는 비교안내문의 내용을 참고하세요",
	// 	"감액 및 특약소멸 감액 또는 특약소멸에 대한 비교안내는 보내드리는 비교안내문을 참고해 주시기 바랍니다",
	// 	"카카오알림톡으로 해피콜을 완료하시면 확인 전화를 드리지 않습니다",
	// 	"휴대폰 전자약관을 선택하셨기 때문에 종이약관은 따로 보내 드리지 않습니다",
	// 	"상품의 중요사항 확인을 위해 해피콜 전화를 드리니 꼭 받아주세요",
	// 	"통신수단 계약해지 동의 통신수단으로 계약 해지 신청이 가능하도록 하시겠습니까?",
	// 	"증권 약관 청약서 등을 휴대폰으로 보내 드릴건데, 동의하시죠?",
	// 	"이메일 전자약관과 종이약관 중, 어떤 걸로 보내 드릴까요?",
	// 	"보내드리는 약관 증권 등 서류의 중요 내용을 확인 부탁 드립니다",
	// 	"약일로부터 사십오일 이내에 청약을 철회할 수 있습니다",
	// 	"청약일로부터 구십일 이내에 청약을 철회할 수 있습니다",
	// 	"이 계약은 예금자보호법에 따라 지급대상 금액을 합하여 일 인당 오천 만 원까지 보호합니다 이해되시죠",
	// 	"이 계약은 예금자보호법에 따라 지급대상 금액을 합하여 일 인당 오천 만 원까지 보호합니다. 이해되시죠?",
	// 	"궁금하신 사항은 고객센터로 문의하시고 분쟁 시에는 금융감독원의 도움을 요청할 수 있습니다",
	// 	"담당 설계사의 불완전판매율 보험계약유지율 등 주요 정보는 이 클린 보험서비스에서 조회하실 수 있습니다",
	// 	"약관 청약서를 받지 못하였거나 약관의 중요한 내용을 설명 받지 못하신 경우 계약성립일로부터 삼 개월 내에 계약 취소가 가능하며 피보험자가 심신상실자이거나 심신박약자로서 사망을 담보하는 계약은 무효입니다",
	// 	"안녕하세요, 저는 씨 제이 오쇼핑 라이나생명 상담사 김미영입니다.",
	// 	"김예스 고객님 맞으신가요?",
	// 	"소중한 시간 내주셔서 감사드리구요.",
	// 	"고객님 요즘 방송에서도 많이 보셨을 텐데요",
	// 	"고객님 연령대에 많이들 걱정하시는게 간병비 보장이시던데 우리가 육십오세 이후에 국민건강보험공단에서 장기요양판정 많이들 받는거 잘 알고 계실 텐데요",
	// 	"그래서 우리가 질병이든 재해든 보통 장기요양등급이다하면 치매만 생각들을 하시는데 치매뿐만이 아니라, 뭐 뇌졸증, 뇌출혈, 파킨슨 중풍 또 각종 재해나 사고로 장기요양 판정을 받는 분들이 많으시거든요",
	// 	"그렇게 국민건강보험 공단에서 장기요양 일 이 삼 사 등급 판정을 받으시면",
	// 	"저희가 최고 사천만원 까지 보장을 받으실 수가 있는데, 내가 만약에 처음부터 일등급을 받게 된다 하시면 총 사천만원을 일시금으로 보장을 받으시고",
	// 	"우선은 사 등급을 판정을 받으셨다 하시면 이천만원 보장 받으시고 보험료 납입 면제가 들어가는데, 증세가 좋아지면 좋겠지만 더 나빠져서 이후에 일등급으로 판정을 받으실수가 있잖아요"}

	// Single Test
	sentenceList := []string{"안녕하세요 처음뵙겠습니다"}

	// Read URL info jsonfile
	b, err := ioutil.ReadFile("./urlinfo.json")
	if err != nil {
		fmt.Println("UrlInfo file read failure")
		fmt.Println(err)
		return
	}
	fmt.Println(b)

	conf := new(urlInfo) // new 로 만들면 포인터
	json.Unmarshal(b, &conf)
	fmt.Printf("%+v\n\n", conf) // show k,v
	fmt.Printf("%#v\n\n", conf)
	fmt.Printf("%s\n\n", conf.APIKey)
	fmt.Printf("%s\n\n", *conf)
	start := time.Now()
	// cnt := 0
	for _, item := range sentenceList {
		wg.Add(1)
		go reqWave(item, &wg, conf)

		// 5개씩 끊는 로직
		// cnt++
		// if cnt%5 == 0 {
		// 	time.Sleep(time.Second * 5)
		// }
		// wg.Wait() // 여기에 wait 을 넣으면 동기 실행됨
	}
	wg.Wait() //  비동기 실행
	elapsed := time.Since(start)
	log.Printf("Total Execution time = %s", elapsed)
}
