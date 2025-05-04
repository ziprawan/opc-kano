package message

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	projectconfig "kano/internals/project_config"
	"kano/internals/utils/kanoutils"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

var wle = []string{"ian", "pdd", "ps:", "htt", "emd", "int", "ikt", "car", "/en", "pi-", "pen", "ll/", "ek.", "i.k", "//a", "c/a", "isa", "id/", "ikt", "go."}

type PDDiktiMahasiswa struct {
	ID          string `json:"id"`
	Nama        string `json:"nama"`
	NIM         string `json:"nim"`
	NamaPT      string `json:"nama_pt"`
	SingkatanPT string `json:"sinkatan_pt"`
	NamaProdi   string `json:"nama_prodi"`
}

type PDDiktiDosen struct {
	ID          string `json:"id"`
	Nama        string `json:"nama"`
	NIDN        string `json:"nidn"`
	NUPTK       string `json:"nuptk"`
	NamaPT      string `json:"nama_pt"`
	SingkatanPT string `json:"sinkatan_pt"`
	NamaProdi   string `json:"nama_prodi"`
}

type PDDiktiPT struct {
	ID          string `json:"id"`
	Kode        string `json:"kode"`
	NamaSingkat string `json:"nama_singkat"`
	Nama        string `json:"nama"`
}

type PDDiktiProdi struct {
	ID        string `json:"id"`
	Nama      string `json:"nama"`
	Jenjang   string `json:"jenjang"`
	PT        string `json:"pt"`
	PTSingkat string `json:"pt_singkat"`
}

type PDDiktiResult struct {
	Mahasiswa []PDDiktiMahasiswa `json:"mahasiswa"`
	Dosen     []PDDiktiDosen     `json:"dosen"`
	PT        []PDDiktiPT        `json:"pt"`
	Prodi     []PDDiktiProdi     `json:"prodi"`
}

func PDDIKTIHandler(ctx *MessageContext) {
	conf := projectconfig.GetConfig()
	if !conf.PDDiktiKey.Valid || !conf.PDDiktiIV.Valid {
		ctx.Instance.Reply("This command is disabled, ask the owner to fix it!", true)
		return
	}

	allowedQueryTypes := []string{"dosen", "mahasiswa", "pt", "prodi", "mhs", "all"}
	fullCmd := ctx.Parser.GetCommand().FullCommand
	args := ctx.Parser.GetArgs()
	if len(args) < 2 {
		ctx.Instance.Reply(fmt.Sprintf("*Pangkalan Data Pendidikan Tinggi (PDDikti) API Wrapper*\nAmbil informasi seputar perguruan tinggi dari PDDikti (yang diambil hanya 3 teratas saja)\n\nPenggunaan:\n%s %s nama/nim\n\nContoh:\n%s dosen Achmad Munir\n%s mahasiswa 2406437994 UI\n%s mhs,prodi,pt nuklir\n%s all telekomunikasi", fullCmd, strings.Join(allowedQueryTypes, ","), fullCmd, fullCmd, fullCmd, fullCmd), true)
		return
	}

	enabledTypes := []bool{false, false, false, false}
	queryType := strings.Split(args[0].Content, ",")
	for _, qType := range queryType {
		idx := slices.Index(allowedQueryTypes, qType)
		if idx == -1 {
			ctx.Instance.Reply(fmt.Sprintf("Tipe kueri yang diizinkan hanya: %s\nDidapat: %s", strings.Join(allowedQueryTypes, ", "), qType), true)
			return
		}

		if idx < 4 {
			enabledTypes[idx] = true
		} else if idx == 4 {
			enabledTypes[1] = true
		} else if idx == 5 {
			enabledTypes[0] = true
			enabledTypes[1] = true
			enabledTypes[2] = true
			enabledTypes[3] = true
		}
	}

	textRunes := []rune(ctx.Parser.Text)
	queryString := url.QueryEscape(string(textRunes[args[1].Start:]))
	url := fmt.Sprintf(
		"%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s",
		wle[3],
		wle[2],
		wle[14],
		wle[9],
		wle[1],
		wle[6],
		wle[13],
		wle[4],
		wle[6],
		wle[16],
		wle[5],
		wle[12],
		wle[19],
		wle[17],
		wle[10],
		wle[7],
		wle[0],
		wle[8],
		wle[15],
		wle[11],
		queryString,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		ctx.Instance.Reply(err.Error(), true)
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ctx.Instance.Reply(err.Error(), true)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		ctx.Instance.Reply(fmt.Sprintf("Expected HTTP code 200, got %s", resp.Status), true)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ctx.Instance.Reply(err.Error(), true)
		return
	}
	cipherText, err := base64.StdEncoding.DecodeString(string(body[1 : len(body)-2]))
	if err != nil {
		ctx.Instance.Reply(fmt.Sprintf("Something went wrong when decoding PDDikti's response\nErr: %s", err), true)
		return
	}
	key, err := base64.StdEncoding.DecodeString(conf.PDDiktiKey.String)
	if err != nil {
		ctx.Instance.Reply(fmt.Sprintf("Something went wrong when decoding PDDikti's key\nErr: %s", err), true)
		return
	}
	iv, err := base64.StdEncoding.DecodeString(conf.PDDiktiIV.String)
	if err != nil {
		ctx.Instance.Reply(fmt.Sprintf("Something went wrong when decoding PDDikti's iv\nErr: %s", err), true)
		return
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		ctx.Instance.Reply(fmt.Sprintf("Something went wrong when initializing new cipher\nErr: %s", err), true)
		return
	}
	if len(cipherText)%aes.BlockSize != 0 {
		ctx.Instance.Reply("CipherText is not a multiple of block size (Maybe encryption or key was changed?)", true)
		return
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(cipherText))
	mode.CryptBlocks(decrypted, cipherText)
	decrypted, err = kanoutils.Pkcs7Unpad(decrypted, aes.BlockSize)
	if err != nil {
		ctx.Instance.Reply(fmt.Sprintf("Something went wrong when unpadding pkcs7\nErr: %s", err), true)
		return
	}

	var result PDDiktiResult
	err = json.Unmarshal(decrypted, &result)
	if err != nil {
		ctx.Instance.Reply(fmt.Sprintf("Something went wrong when parsing decrypted result (maybe the structure was changed?)\nErr: %s", err), true)
		return
	}

	// Create msg
	msgs := []string{}

	if len(result.Dosen) > 0 && enabledTypes[0] {
		adds := "*List Dosen*\n==========\n\n"
		dosens := []string{}

		for idx, dosen := range result.Dosen {
			if idx >= 3 {
				break
			}
			dosenAdds := fmt.Sprintf("Nama: %s\n", dosen.Nama)
			dosenAdds += fmt.Sprintf("Nomor Induk Dosen Nasional: %s\n", dosen.NIDN)
			if dosen.NUPTK != "" {
				dosenAdds += fmt.Sprintf("Nomor Unik Pendidik dan Tenaga Kependidikan : %s\n", dosen.NUPTK)
			}
			dosenAdds += fmt.Sprintf("Perguruan Tinggi: %s", dosen.NamaPT)
			if dosen.SingkatanPT != "" {
				dosenAdds += fmt.Sprintf(" (%s)", dosen.SingkatanPT)
			}
			dosenAdds += fmt.Sprintf("\nProgram Studi: %s", dosen.NamaProdi)

			dosens = append(dosens, dosenAdds)
		}

		adds += strings.Join(dosens, "\n----------\n")
		msgs = append(msgs, adds)
	}
	if len(result.Mahasiswa) > 0 && enabledTypes[1] {
		adds := "*List Mahasiswa*\n==========\n\n"
		mhss := []string{}

		for idx, mhs := range result.Mahasiswa {
			if idx >= 3 {
				break
			}
			mhsAdds := fmt.Sprintf("Nama: %s\n", mhs.Nama)
			mhsAdds += fmt.Sprintf("Nomor Induk Mahasiswa: %s\n", mhs.NIM)
			mhsAdds += fmt.Sprintf("Perguruan Tinggi: %s", mhs.NamaPT)
			if mhs.SingkatanPT != "" {
				mhsAdds += fmt.Sprintf(" (%s)", mhs.SingkatanPT)
			}
			mhsAdds += fmt.Sprintf("\nProgram Studi: %s", mhs.NamaProdi)

			mhss = append(mhss, mhsAdds)
		}

		adds += strings.Join(mhss, "\n----------\n")
		msgs = append(msgs, adds)
	}
	if len(result.PT) > 0 && enabledTypes[2] {
		adds := "*List Perguruan Tinggi*\n==========\n\n"
		pts := []string{}

		for idx, pt := range result.PT {
			if idx >= 3 {
				break
			}
			ptAdds := fmt.Sprintf("Nama: %s", pt.Nama)
			if pt.NamaSingkat != "" {
				ptAdds += fmt.Sprintf(" (%s)", pt.NamaSingkat)
			}
			ptAdds += fmt.Sprintf("\nKode: %s", pt.Kode)

			pts = append(pts, ptAdds)
		}

		adds += strings.Join(pts, "\n----------\n")
		msgs = append(msgs, adds)
	}
	if len(result.Prodi) > 0 && enabledTypes[3] {
		adds := "*List Program Studi*\n==========\n\n"
		prodis := []string{}

		for idx, prodi := range result.Prodi {
			if idx >= 3 {
				break
			}
			mhsAdds := fmt.Sprintf("Nama: %s - %s\nPerguruan Tinggi: %s", prodi.Jenjang, prodi.Nama, prodi.PT)
			if prodi.PTSingkat != "" {
				mhsAdds += fmt.Sprintf(" (%s)", prodi.PTSingkat)
			}

			prodis = append(prodis, mhsAdds)
		}

		adds += strings.Join(prodis, "\n----------\n")
		msgs = append(msgs, adds)
	}

	if len(msgs) == 0 {
		ctx.Instance.Reply("Server returned an empty data", true)
	} else {
		ctx.Instance.Reply(strings.Join(msgs, "\n\n"), true)
	}
}
