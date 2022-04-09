package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

const (
	rootDir = "contracts"

	NounsTokenAddress                  = "0x9C8fF314C9Bc7F6e59A9d9225Fb22946427eDC03"
	NounsSeederAddress                 = "0xCC8a0FB5ab3C7132c1b2A0109142Fb112c4Ce515"
	NounsDescriptorAddress             = "0x0Cfdb3Ba1694c2bb2CFACB0339ad7b1Ae5932B63"
	NFTDescriptorAddress               = "0x0BBAd8c947210ab6284699605ce2a61780958264"
	NounsAuctionHouseAddress           = "0xF15a943787014461d94da08aD4040f79Cd7c124e"
	NounsNounsAuctionHouseProxyAddress = "0x830BD73E4184ceF73443C15111a1DF14e495C706"
	NounsAuctionHouseProxyAdminAddress = "0xC1C119932d78aB9080862C5fcb964029f086401e"
	NounsDAOExecutorAddress            = "0x0BC3807Ec262cB779b38D65b38158acC3bfedE10"
	NounsDAOProxyAddress               = "0x6f3E6272A167e8AcCb32072d08E0957F9c79223d"
	NounsDAOLogicV1Address             = "0xa43aFE317985726E4e194eb061Af77fbCb43F944"

	NounsTokenDir                  = "nouns_token"
	NounsSeederDir                 = "nouns_seeder"
	NounsDescriptorDir             = "nouns_descriptor"
	NFTDescriptorDir               = "nft_descriptor_dir"
	NounsAuctionHouseDir           = "nouns_auction_house"
	NounsNounsAuctionHouseProxyDir = "nouns_nouns_auction_house_proxy"
	NounsAuctionHouseProxyAdminDir = "nouns_auction_house_proxy_admin"
	NounsDAOExecutorDir            = "nouns_dao_executor"
	NounsDAOProxyDir               = "nouns_dao_proxy"
	NounsDAOLogicV1Dir             = "nouns_dao_logic_v1"
)

var addressToDir = map[string]string{
	NounsTokenAddress:                  NounsTokenDir,
	NounsSeederAddress:                 NounsSeederDir,
	NounsDescriptorAddress:             NounsDescriptorDir,
	NFTDescriptorAddress:               NFTDescriptorDir,
	NounsAuctionHouseAddress:           NounsAuctionHouseDir,
	NounsNounsAuctionHouseProxyAddress: NounsNounsAuctionHouseProxyDir,
	NounsAuctionHouseProxyAdminAddress: NounsAuctionHouseProxyAdminDir,
	NounsDAOExecutorAddress:            NounsDAOExecutorDir,
	NounsDAOProxyAddress:               NounsDAOProxyDir,
	NounsDAOLogicV1Address:             NounsDAOLogicV1Dir,
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	for address, dir := range addressToDir {
		if err := getContractSources(address, filepath.Join(rootDir, dir)); err != nil {
			return err
		}
	}

	return nil
}

func getContractSources(address, dir string) error {
	rawCodes, err := getRawContractCode(address)
	if err != nil {
		return err
	}

	sourceCodes, err := parseContractCode(rawCodes)
	if err != nil {
		return err
	}

	for _, sourceCode := range sourceCodes {
		for path, source := range sourceCode.Sources {
			if err := os.MkdirAll(targetDir(dir, path), os.ModePerm); err != nil {
				return err
			}

			f, err := os.Create(targetPath(dir, path))
			if err != nil {
				return err
			}
			defer f.Close()

			f.Write([]byte(source.Content))
		}
	}

	return nil
}

func getContractURL(address string, apikey string) string {
	const url = "https://api.etherscan.io/api?module=contract&action=getsourcecode&address=%s&apikey=%s"
	return fmt.Sprintf(url, address, apikey)
}

func targetDir(dir string, path string) string {
	return filepath.Dir(targetPath(dir, path))
}

func targetPath(dir string, path string) string {
	return filepath.Join(dir, path)
}

func getRawContractCode(address string) ([]*RawCode, error) {
	url := getContractURL(address, os.Getenv("ETHERSCAN_APIKEY"))
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, err
	}

	contractCodeResponse := &Response{}

	if err := json.NewDecoder(resp.Body).Decode(contractCodeResponse); err != nil {
		return nil, err
	}

	if contractCodeResponse.Status != "1" {
		return nil, fmt.Errorf("bad status: %s, message: %s", contractCodeResponse.Status, contractCodeResponse.Status)
	}

	return contractCodeResponse.Codes, nil
}

func parseContractCode(rawCodes []*RawCode) ([]*SourceCode, error) {
	sourceCodes := make([]*SourceCode, 0, len(rawCodes))
	for _, rawCode := range rawCodes {
		sourceCode := &SourceCode{}
		if err := json.Unmarshal([]byte(rawCode.SourceCode[1:len(rawCode.SourceCode)-1]), sourceCode); err != nil {
			return nil, err
		}

		sourceCodes = append(sourceCodes, sourceCode)
	}

	return sourceCodes, nil
}

type Response struct {
	Status  string     `json:"status"`
	Message string     `json:"message"`
	Codes   []*RawCode `json:"result"`
}

type RawCode struct {
	SourceCode           string `json:"SourceCode"`
	Abi                  string `json:"ABI"`
	ContractName         string `json:"ContractName"`
	CompilerVersion      string `json:"CompilerVersion"`
	OptimizationUsed     string `json:"OptimizationUsed"`
	Runs                 string `json:"Runs"`
	ConstructorArguments string `json:"ConstructorArguments"`
	EVMVersion           string `json:"EVMVersion"`
	Library              string `json:"Library"`
	LicenseType          string `json:"LicenseType"`
	Proxy                string `json:"Proxy"`
	Implementation       string `json:"Implementation"`
	SwarmSource          string `json:"SwarmSource"`
}

// SourceCodeFields
type SourceCode struct {
	Language string   `json:"language"`
	Sources  Sources  `json:"sources"`
	Settings Settings `json:"settings"`
}

type Sources map[string]*Contract

type Contract struct {
	Content string `json:"content"`
}

type Settings struct {
	Optimizer       *Optimizer      `json:"optimizer"`
	OutputSelection OutputSelection `json:"outputSelection"`
	Libraries       Libraries       `json:"libraries"`
}

type Optimizer struct {
	Enabled bool `json:"enabled"`
	Runs    int  `json:"runs"`
}

type OutputSelection map[string]map[string][]string

type Libraries interface{}
