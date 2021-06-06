package structs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adjust/rmq/v3"
	"github.com/go-pg/pg/v10"
	"strconv"
	"time"
)

type DBLogger struct{}

func (d DBLogger) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
	return c, nil
}

func (d DBLogger) AfterQuery(c context.Context, q *pg.QueryEvent) error {
	fmt.Println(q.FormattedQuery())
	return nil
}

// DBParams struct for storing the postgres db's details
type DBParams struct {
	User     string
	Password string
	Dbname   string
	Host     string
	Port     string
}

// structs for response from the above http query
type (
	AggV2 struct {
		V  float64 `json:"v"`
		Vw float64 `json:"vw"`
		O  float64 `json:"o"`
		C  float64 `json:"c"`
		H  float64 `json:"h"`
		L  float64 `json:"l"`
		T  float64 `json:"t"`
		N  float64 `json:"n"`
	}
	ResponseParams struct {
		Ticker       string  `json:"ticker"`
		Status       string  `json:"status"`
		QueryCount   int     `json:"queryCount"`
		ResultsCount int     `json:"resultsCount"`
		Adjusted     bool    `json:"adjusted"`
		Results      []AggV2 `json:"results"`
		RequestId    string  `json:"request_id"`
	}
	ExpandedResponseParams struct {
		Ticker     string
		Timespan   string
		Multiplier int
		V          float64
		Vw         float64
		O          float64
		C          float64
		H          float64
		L          float64
		T          string
	}
	RequestsURL struct {
		Url string `json:"url"`
	}
)

// for reading the json file 'responses.json', we need to have a super structure of all the other queries
type (
	Part struct {
		Request  RequestsURL    `json:"request"`
		Response ResponseParams `json:"response"`
	}
	ResponsesJSONFile struct {
		StocksAgg Part `json:"aggregates"`
	}
)

// tickers
type (
	Codes struct {
		Cik     string `json:"cik"`
		Figiuid string `json:"figiuid"`
		Scfigi  string `json:"scfigi"`
		Cfigi   string `json:"cfigi"`
		Figi    string `json:"figi"`
	}
	Attrs struct {
		Currencyname string `json:"currencyName"`
		Currency     string `json:"currency"`
		Basename     string `json:"baseName"`
		Base         string `json:"base"`
	}
	TickersInner struct {
		Ticker      string    `json:"ticker"`
		Name        string    `json:"name"`
		Market      string    `json:"market"`
		Locale      string    `json:"locale"`
		Currency    string    `json:"currency"`
		Active      bool      `json:"active"`
		Primaryexch string    `json:"primaryExch"`
		Type        *string   `json:"type,omitempty"`
		Codes       *Codes    `json:"codes,omitempty"`
		Updated     time.Time `json:"updated"`
		URL         string    `json:"url"`
		Attrs       *Attrs    `json:"attrs,omitempty"`
	}
	TickersResponse struct {
		Page    int            `json:"page"`
		Perpage int            `json:"perPage"`
		Count   int            `json:"count"`
		Status  string         `json:"status"`
		Tickers []TickersInner `json:"tickers"`
	}
	Tickers struct {
		InsertDate   time.Time `json:"insert_date"`
		Page         int       `json:"page"`
		Perpage      int       `json:"perPage"`
		Count        int       `json:"count"`
		Status       string    `json:"status"`
		Ticker       string    `json:"ticker"`
		Name         string    `json:"name"`
		Market       string    `json:"market"`
		Locale       string    `json:"locale"`
		Currency     string    `json:"currency"`
		Active       bool      `json:"active"`
		Primaryexch  string    `json:"primaryExch"`
		Type         *string   `json:"type,omitempty"`
		Cik          string    `json:"cik"`
		Figiuid      string    `json:"figiuid"`
		Scfigi       string    `json:"scfigi"`
		Cfigi        string    `json:"cfigi"`
		Figi         string    `json:"figi"`
		Updated      time.Time `json:"updated"`
		URL          string    `json:"url"`
		Currencyname string    `json:"currencyName"`
		Basename     string    `json:"baseName"`
		Base         string    `json:"base"`
	}
)

// tickers VX
type (
	TickerVxInnerResponse struct {
		Ticker          string `json:"ticker"`
		Name            string `json:"name"`
		Market          string `json:"market"`
		Locale          string `json:"locale"`
		PrimaryExchange string `json:"primary_exchange"`
		Type            string `json:"type"`
		Active          bool   `json:"active"`
		CurrencyName    string `json:"currency_name"`
		Cik             string `json:"cik"`
		CompositeFigi   string `json:"composite_figi"`
		ShareClassFigi  string `json:"share_class_figi"`
		LastUpdatedUtc  string `json:"last_updated_utc"`
	}
	TickersVxResponse struct {
		Results   []TickerVxInnerResponse `json:"results"`
		Status    string                  `json:"status"`
		RequestId string                  `json:"request_id"`
		Count     int64                   `json:"count"`
		NextUrl   string                  `json:"next_url"`
	}
	TickerVx struct {
		InsertDatetime  time.Time `json:"insert_datetime"`
		Ticker          string    `json:"ticker"`
		Name            string    `json:"name"`
		Market          string    `json:"market"`
		Locale          string    `json:"locale"`
		PrimaryExchange string    `json:"primary_exchange"`
		Type            string    `json:"type"`
		Active          bool      `json:"active"`
		CurrencyName    string    `json:"currency_name"`
		Cik             string    `json:"cik"`
		CompositeFigi   string    `json:"composite_figi"`
		ShareClassFigi  string    `json:"share_class_figi"`
		LastUpdatedUtc  string    `json:"last_updated_utc"`
	}
)

// ticker types
type (
	TypesResponse struct {
		Cs    string `json:"CS"`
		Adr   string `json:"ADR"`
		Nvdr  string `json:"NVDR"`
		Gdr   string `json:"GDR"`
		Sdr   string `json:"SDR"`
		Cef   string `json:"CEF"`
		Etp   string `json:"ETP"`
		Reit  string `json:"REIT"`
		Mlp   string `json:"MLP"`
		Wrt   string `json:"WRT"`
		Pub   string `json:"PUB"`
		Nyrs  string `json:"NYRS"`
		Unit  string `json:"UNIT"`
		Right string `json:"RIGHT"`
		Track string `json:"TRACK"`
		Ltdp  string `json:"LTDP"`
		Rylt  string `json:"RYLT"`
		Mf    string `json:"MF"`
		Pfd   string `json:"PFD"`
		Fdr   string `json:"FDR"`
		Ost   string `json:"OST"`
		Fund  string `json:"FUND"`
		Sp    string `json:"SP"`
		Si    string `json:"SI"`
	}
	IndexTypesResponse struct {
		Index      string `json:"INDEX"`
		Etf        string `json:"ETF"`
		Etn        string `json:"ETN"`
		Etmf       string `json:"ETMF"`
		Settlement string `json:"SETTLEMENT"`
		Spot       string `json:"SPOT"`
		Subprod    string `json:"SUBPROD"`
		Wc         string `json:"WC"`
		Alphaindex string `json:"ALPHAINDEX"`
	}
	TotalResponse struct {
		Types      TypesResponse      `json:"types"`
		IndexTypes IndexTypesResponse `json:"indexTypes"`
	}
	TickerTypeResponse struct {
		Status  string        `json:"status"`
		Results TotalResponse `json:"results"`
	}
	TickerType struct {
		Cs         string `json:"CS"`
		Adr        string `json:"ADR"`
		Nvdr       string `json:"NVDR"`
		Gdr        string `json:"GDR"`
		Sdr        string `json:"SDR"`
		Cef        string `json:"CEF"`
		Etp        string `json:"ETP"`
		Reit       string `json:"REIT"`
		Mlp        string `json:"MLP"`
		Wrt        string `json:"WRT"`
		Pub        string `json:"PUB"`
		Nyrs       string `json:"NYRS"`
		Unit       string `json:"UNIT"`
		Right      string `json:"RIGHT"`
		Track      string `json:"TRACK"`
		Ltdp       string `json:"LTDP"`
		Rylt       string `json:"RYLT"`
		Mf         string `json:"MF"`
		Pfd        string `json:"PFD"`
		Fdr        string `json:"FDR"`
		Ost        string `json:"OST"`
		Fund       string `json:"FUND"`
		Sp         string `json:"SP"`
		Si         string `json:"SI"`
		Index      string `json:"INDEX"`
		Etf        string `json:"ETF"`
		Etn        string `json:"ETN"`
		Etmf       string `json:"ETMF"`
		Settlement string `json:"SETTLEMENT"`
		Spot       string `json:"SPOT"`
		Subprod    string `json:"SUBPROD"`
		Wc         string `json:"WC"`
		Alphaindex string `json:"ALPHAINDEX"`
	}
)

// TickerDetails
type TickerDetails struct {
	Logo           string   `json:"logo"`
	Listdate       string   `json:"listdate"`
	Cik            string   `json:"cik"`
	Figi           string   `json:"figi"`
	Lei            string   `json:"lei"`
	Sic            string   `json:"sic"`
	Country        string   `json:"country"`
	Industry       string   `json:"industry"`
	Sector         string   `json:"sector"`
	Marketcap      float64  `json:"marketcap"`
	Employees      string   `json:"employees"`
	Phone          string   `json:"phone"`
	Ceo            string   `json:"ceo"`
	Url            string   `json:"url"`
	Description    string   `json:"description"`
	Exchange       string   `json:"exchange"`
	Name           string   `json:"name"`
	Symbol         string   `json:"symbol"`
	ExchangeSymbol string   `json:"ExchangeSymbol"`
	HqAddress      string   `json:"hq_address"`
	HqState        string   `json:"hq_state"`
	HqCountry      string   `json:"hq_country"`
	Type           string   `json:"type"`
	Updated        string   `json:"updated"`
	Tags           []string `json:"tags"`
	Similar        []string `json:"similar"`
	Active         bool     `json:"active"`
}

// TickerNews
type TickerNews struct {
	Symbols   []string `json:"symbols"`
	Timestamp string   `json:"timestamp"`
	Title     string   `json:"title"`
	Url       string   `json:"url"`
	Source    string   `json:"source"`
	Summary   string   `json:"summary"`
	Image     string   `json:"image"`
	Keywords  []string `json:"keywords"`
}

// markets
type (
	MarketsResultsResponse struct {
		Market string `json:"market"`
		Desc   string `json:"desc"`
	}
	MarketsResponse struct {
		Status  string                   `json:"status"`
		Results []MarketsResultsResponse `json:"results"`
	}
	Markets struct {
		InsertDate time.Time `json:"insert_date"`
		Status     string    `json:"status"`
		Market     string    `json:"market"`
		Desc       string    `json:"desc"`
	}
)

// locales
type (
	LocalesResultsResponse struct {
		Locals string `json:"locals"`
		Name   string `json:"name"`
	}
	LocalesResponse struct {
		Status  string                   `json:"status"`
		Results []LocalesResultsResponse `json:"results"`
	}
	Locales struct {
		InsertDate time.Time `json:"insert_date"`
		Status     string    `json:"status"`
		Locals     string    `json:"locals"`
		Name       string    `json:"name"`
	}
)

// stock splits
type (
	StockSplitsResultsResponse struct {
		Ticker       string  `json:"ticker"`
		ExDate       string  `json:"ExDate"`
		PaymentDate  string  `json:"PaymentDate"`
		DeclaredDate string  `json:"DeclaredDate,omitempty"`
		Ratio        float64 `json:"ratio"`
		ToFactor     float64 `json:"ToFactor"`
		ForFactor    float64 `json:"ForFactor"`
	}
	StockSplitsResponse struct {
		Status  string                       `json:"status"`
		Count   string                       `json:"count"`
		Results []StockSplitsResultsResponse `json:"results"`
	}
	StockSplits struct {
		InsertDatetime time.Time `json:"insert_datetime"`
		Status         string    `json:"status"`
		Count          string    `json:"count"`
		Ticker         string    `json:"ticker"`
		ExDate         string    `json:"ExDate"`
		PaymentDate    string    `json:"PaymentDate"`
		DeclaredDate   string    `json:"DeclaredDate,omitempty"`
		Ratio          float64   `json:"ratio"`
		ToFactor       float64   `json:"ToFactor"`
		ForFactor      float64   `json:"ForFactor"`
	}
)

// StockDividendsResponse
type StockDividendsResponse struct {
	Status  string `json:"status"`
	Count   int    `json:"count"`
	Results []struct {
		Ticker       string  `json:"ticker"`
		Exdate       string  `json:"exDate"`
		Paymentdate  string  `json:"paymentDate"`
		Recorddate   string  `json:"recordDate"`
		Amount       float64 `json:"amount"`
		Declareddate string  `json:"declaredDate,omitempty"`
	} `json:"results"`
}

// StockDividends
type StockDividends struct {
	InsertDate   time.Time `json:"insert_date"`
	Status       string    `json:"status"`
	Count        int       `json:"count"`
	Ticker       string    `json:"ticker"`
	Exdate       string    `json:"exDate"`
	Paymentdate  string    `json:"paymentDate"`
	Recorddate   string    `json:"recordDate"`
	Amount       float64   `json:"amount"`
	Declareddate string    `json:"declaredDate,omitempty"`
}

// StockFinancialsResponse
type StockFinancialsResponse struct {
	Status  string `json:"status"`
	Results []struct {
		Ticker                                                 string  `json:"ticker"`
		Period                                                 string  `json:"period"`
		Calendardate                                           string  `json:"calendarDate"`
		Reportperiod                                           string  `json:"reportPeriod"`
		Updated                                                string  `json:"updated"`
		Datekey                                                string  `json:"dateKey"`
		Accumulatedothercomprehensiveincome                    int64   `json:"accumulatedOtherComprehensiveIncome"`
		Assets                                                 int64   `json:"assets"`
		Assetscurrent                                          int64   `json:"assetsCurrent"`
		Assetsnoncurrent                                       int64   `json:"assetsNonCurrent"`
		Bookvaluepershare                                      float64 `json:"bookValuePerShare"`
		Capitalexpenditure                                     int     `json:"capitalExpenditure"`
		Cashandequivalents                                     int64   `json:"cashAndEquivalents"`
		Cashandequivalentsusd                                  int64   `json:"cashAndEquivalentsUSD"`
		Costofrevenue                                          int64   `json:"costOfRevenue"`
		Consolidatedincome                                     int64   `json:"consolidatedIncome"`
		Currentratio                                           float64 `json:"currentRatio"`
		Debttoequityratio                                      float64 `json:"debtToEquityRatio"`
		Debt                                                   int64   `json:"debt"`
		Debtcurrent                                            int64   `json:"debtCurrent"`
		Debtnoncurrent                                         int64   `json:"debtNonCurrent"`
		Debtusd                                                int64   `json:"debtUSD"`
		Deferredrevenue                                        int64   `json:"deferredRevenue"`
		Depreciationamortizationandaccretion                   int64   `json:"depreciationAmortizationAndAccretion"`
		Deposits                                               int     `json:"deposits"`
		Dividendyield                                          float64 `json:"dividendYield"`
		Dividendsperbasiccommonshare                           float64 `json:"dividendsPerBasicCommonShare"`
		Earningbeforeinteresttaxes                             int64   `json:"earningBeforeInterestTaxes"`
		Earningsbeforeinteresttaxesdepreciationamortization    int64   `json:"earningsBeforeInterestTaxesDepreciationAmortization"`
		Ebitdamargin                                           float64 `json:"EBITDAMargin"`
		Earningsbeforeinteresttaxesdepreciationamortizationusd int64   `json:"earningsBeforeInterestTaxesDepreciationAmortizationUSD"`
		Earningbeforeinteresttaxesusd                          int64   `json:"earningBeforeInterestTaxesUSD"`
		Earningsbeforetax                                      int64   `json:"earningsBeforeTax"`
		Earningsperbasicshare                                  float64 `json:"earningsPerBasicShare"`
		Earningsperdilutedshare                                float64 `json:"earningsPerDilutedShare"`
		Earningsperbasicshareusd                               float64 `json:"earningsPerBasicShareUSD"`
		Shareholdersequity                                     int64   `json:"shareholdersEquity"`
		Shareholdersequityusd                                  int64   `json:"shareholdersEquityUSD"`
		Enterprisevalue                                        int64   `json:"enterpriseValue"`
		Enterprisevalueoverebit                                int     `json:"enterpriseValueOverEBIT"`
		Enterprisevalueoverebitda                              float64 `json:"enterpriseValueOverEBITDA"`
		Freecashflow                                           int64   `json:"freeCashFlow"`
		Freecashflowpershare                                   float64 `json:"freeCashFlowPerShare"`
		Foreigncurrencyusdexchangerate                         int     `json:"foreignCurrencyUSDExchangeRate"`
		Grossprofit                                            int64   `json:"grossProfit"`
		Grossmargin                                            float64 `json:"grossMargin"`
		Goodwillandintangibleassets                            int     `json:"goodwillAndIntangibleAssets"`
		Interestexpense                                        int     `json:"interestExpense"`
		Investedcapital                                        int64   `json:"investedCapital"`
		Inventory                                              int64   `json:"inventory"`
		Investments                                            int64   `json:"investments"`
		Investmentscurrent                                     int64   `json:"investmentsCurrent"`
		Investmentsnoncurrent                                  int64   `json:"investmentsNonCurrent"`
		Totalliabilities                                       int64   `json:"totalLiabilities"`
		Currentliabilities                                     int64   `json:"currentLiabilities"`
		Liabilitiesnoncurrent                                  int64   `json:"liabilitiesNonCurrent"`
		Marketcapitalization                                   int64   `json:"marketCapitalization"`
		Netcashflow                                            int     `json:"netCashFlow"`
		Netcashflowbusinessacquisitionsdisposals               int     `json:"netCashFlowBusinessAcquisitionsDisposals"`
		Issuanceequityshares                                   int64   `json:"issuanceEquityShares"`
		Issuancedebtsecurities                                 int     `json:"issuanceDebtSecurities"`
		Paymentdividendsothercashdistributions                 int64   `json:"paymentDividendsOtherCashDistributions"`
		Netcashflowfromfinancing                               int64   `json:"netCashFlowFromFinancing"`
		Netcashflowfrominvesting                               int64   `json:"netCashFlowFromInvesting"`
		Netcashflowinvestmentacquisitionsdisposals             int64   `json:"netCashFlowInvestmentAcquisitionsDisposals"`
		Netcashflowfromoperations                              int64   `json:"netCashFlowFromOperations"`
		Effectofexchangeratechangesoncash                      int     `json:"effectOfExchangeRateChangesOnCash"`
		Netincome                                              int64   `json:"netIncome"`
		Netincomecommonstock                                   int64   `json:"netIncomeCommonStock"`
		Netincomecommonstockusd                                int64   `json:"netIncomeCommonStockUSD"`
		Netlossincomefromdiscontinuedoperations                int     `json:"netLossIncomeFromDiscontinuedOperations"`
		Netincometononcontrollinginterests                     int     `json:"netIncomeToNonControllingInterests"`
		Profitmargin                                           float64 `json:"profitMargin"`
		Operatingexpenses                                      int64   `json:"operatingExpenses"`
		Operatingincome                                        int64   `json:"operatingIncome"`
		Tradeandnontradepayables                               int64   `json:"tradeAndNonTradePayables"`
		Payoutratio                                            float64 `json:"payoutRatio"`
		Pricetobookvalue                                       float64 `json:"priceToBookValue"`
		Priceearnings                                          float64 `json:"priceEarnings"`
		Pricetoearningsratio                                   float64 `json:"priceToEarningsRatio"`
		Propertyplantequipmentnet                              int64   `json:"propertyPlantEquipmentNet"`
		Preferreddividendsincomestatementimpact                int     `json:"preferredDividendsIncomeStatementImpact"`
		Sharepriceadjustedclose                                float64 `json:"sharePriceAdjustedClose"`
		Pricesales                                             float64 `json:"priceSales"`
		Pricetosalesratio                                      float64 `json:"priceToSalesRatio"`
		Tradeandnontradereceivables                            int64   `json:"tradeAndNonTradeReceivables"`
		Accumulatedretainedearningsdeficit                     int64   `json:"accumulatedRetainedEarningsDeficit"`
		Revenues                                               int64   `json:"revenues"`
		Revenuesusd                                            int64   `json:"revenuesUSD"`
		Researchanddevelopmentexpense                          int64   `json:"researchAndDevelopmentExpense"`
		Sharebasedcompensation                                 int     `json:"shareBasedCompensation"`
		Sellinggeneralandadministrativeexpense                 int64   `json:"sellingGeneralAndAdministrativeExpense"`
		Sharefactor                                            int     `json:"shareFactor"`
		Shares                                                 int64   `json:"shares"`
		Weightedaverageshares                                  int64   `json:"weightedAverageShares"`
		Weightedaveragesharesdiluted                           int64   `json:"weightedAverageSharesDiluted"`
		Salespershare                                          float64 `json:"salesPerShare"`
		Tangibleassetvalue                                     int64   `json:"tangibleAssetValue"`
		Taxassets                                              int     `json:"taxAssets"`
		Incometaxexpense                                       int     `json:"incomeTaxExpense"`
		Taxliabilities                                         int     `json:"taxLiabilities"`
		Tangibleassetsbookvaluepershare                        float64 `json:"tangibleAssetsBookValuePerShare"`
		Workingcapital                                         int64   `json:"workingCapital"`
		Assetsaverage                                          int64   `json:"assetsAverage,omitempty"`
		Assetturnover                                          float64 `json:"assetTurnover,omitempty"`
		Averageequity                                          int64   `json:"averageEquity,omitempty"`
		Investedcapitalaverage                                 int64   `json:"investedCapitalAverage,omitempty"`
		Returnonaverageassets                                  float64 `json:"returnOnAverageAssets,omitempty"`
		Returnonaverageequity                                  float64 `json:"returnOnAverageEquity,omitempty"`
		Returnoninvestedcapital                                float64 `json:"returnOnInvestedCapital,omitempty"`
		Returnonsales                                          float64 `json:"returnOnSales,omitempty"`
	} `json:"results"`
}

// StockFinancials
type StockFinancials struct {
	Status                                                 string    `json:"status"`
	InsertDate                                             time.Time `json:"insert_date"`
	Ticker                                                 string    `json:"ticker"`
	Period                                                 string    `json:"period"`
	Calendardate                                           string    `json:"calendarDate"`
	Reportperiod                                           string    `json:"reportPeriod"`
	Updated                                                string    `json:"updated"`
	Datekey                                                string    `json:"dateKey"`
	Accumulatedothercomprehensiveincome                    int64     `json:"accumulatedOtherComprehensiveIncome"`
	Assets                                                 int64     `json:"assets"`
	Assetscurrent                                          int64     `json:"assetsCurrent"`
	Assetsnoncurrent                                       int64     `json:"assetsNonCurrent"`
	Bookvaluepershare                                      float64   `json:"bookValuePerShare"`
	Capitalexpenditure                                     int       `json:"capitalExpenditure"`
	Cashandequivalents                                     int64     `json:"cashAndEquivalents"`
	Cashandequivalentsusd                                  int64     `json:"cashAndEquivalentsUSD"`
	Costofrevenue                                          int64     `json:"costOfRevenue"`
	Consolidatedincome                                     int64     `json:"consolidatedIncome"`
	Currentratio                                           float64   `json:"currentRatio"`
	Debttoequityratio                                      float64   `json:"debtToEquityRatio"`
	Debt                                                   int64     `json:"debt"`
	Debtcurrent                                            int64     `json:"debtCurrent"`
	Debtnoncurrent                                         int64     `json:"debtNonCurrent"`
	Debtusd                                                int64     `json:"debtUSD"`
	Deferredrevenue                                        int64     `json:"deferredRevenue"`
	Depreciationamortizationandaccretion                   int64     `json:"depreciationAmortizationAndAccretion"`
	Deposits                                               int       `json:"deposits"`
	Dividendyield                                          float64   `json:"dividendYield"`
	Dividendsperbasiccommonshare                           float64   `json:"dividendsPerBasicCommonShare"`
	Earningbeforeinteresttaxes                             int64     `json:"earningBeforeInterestTaxes"`
	Earningsbeforeinteresttaxesdepreciationamortization    int64     `json:"earningsBeforeInterestTaxesDepreciationAmortization"`
	Ebitdamargin                                           float64   `json:"EBITDAMargin"`
	Earningsbeforeinteresttaxesdepreciationamortizationusd int64     `json:"earningsBeforeInterestTaxesDepreciationAmortizationUSD"`
	Earningbeforeinteresttaxesusd                          int64     `json:"earningBeforeInterestTaxesUSD"`
	Earningsbeforetax                                      int64     `json:"earningsBeforeTax"`
	Earningsperbasicshare                                  float64   `json:"earningsPerBasicShare"`
	Earningsperdilutedshare                                float64   `json:"earningsPerDilutedShare"`
	Earningsperbasicshareusd                               float64   `json:"earningsPerBasicShareUSD"`
	Shareholdersequity                                     int64     `json:"shareholdersEquity"`
	Shareholdersequityusd                                  int64     `json:"shareholdersEquityUSD"`
	Enterprisevalue                                        int64     `json:"enterpriseValue"`
	Enterprisevalueoverebit                                int       `json:"enterpriseValueOverEBIT"`
	Enterprisevalueoverebitda                              float64   `json:"enterpriseValueOverEBITDA"`
	Freecashflow                                           int64     `json:"freeCashFlow"`
	Freecashflowpershare                                   float64   `json:"freeCashFlowPerShare"`
	Foreigncurrencyusdexchangerate                         int       `json:"foreignCurrencyUSDExchangeRate"`
	Grossprofit                                            int64     `json:"grossProfit"`
	Grossmargin                                            float64   `json:"grossMargin"`
	Goodwillandintangibleassets                            int       `json:"goodwillAndIntangibleAssets"`
	Interestexpense                                        int       `json:"interestExpense"`
	Investedcapital                                        int64     `json:"investedCapital"`
	Inventory                                              int64     `json:"inventory"`
	Investments                                            int64     `json:"investments"`
	Investmentscurrent                                     int64     `json:"investmentsCurrent"`
	Investmentsnoncurrent                                  int64     `json:"investmentsNonCurrent"`
	Totalliabilities                                       int64     `json:"totalLiabilities"`
	Currentliabilities                                     int64     `json:"currentLiabilities"`
	Liabilitiesnoncurrent                                  int64     `json:"liabilitiesNonCurrent"`
	Marketcapitalization                                   int64     `json:"marketCapitalization"`
	Netcashflow                                            int       `json:"netCashFlow"`
	Netcashflowbusinessacquisitionsdisposals               int       `json:"netCashFlowBusinessAcquisitionsDisposals"`
	Issuanceequityshares                                   int64     `json:"issuanceEquityShares"`
	Issuancedebtsecurities                                 int       `json:"issuanceDebtSecurities"`
	Paymentdividendsothercashdistributions                 int64     `json:"paymentDividendsOtherCashDistributions"`
	Netcashflowfromfinancing                               int64     `json:"netCashFlowFromFinancing"`
	Netcashflowfrominvesting                               int64     `json:"netCashFlowFromInvesting"`
	Netcashflowinvestmentacquisitionsdisposals             int64     `json:"netCashFlowInvestmentAcquisitionsDisposals"`
	Netcashflowfromoperations                              int64     `json:"netCashFlowFromOperations"`
	Effectofexchangeratechangesoncash                      int       `json:"effectOfExchangeRateChangesOnCash"`
	Netincome                                              int64     `json:"netIncome"`
	Netincomecommonstock                                   int64     `json:"netIncomeCommonStock"`
	Netincomecommonstockusd                                int64     `json:"netIncomeCommonStockUSD"`
	Netlossincomefromdiscontinuedoperations                int       `json:"netLossIncomeFromDiscontinuedOperations"`
	Netincometononcontrollinginterests                     int       `json:"netIncomeToNonControllingInterests"`
	Profitmargin                                           float64   `json:"profitMargin"`
	Operatingexpenses                                      int64     `json:"operatingExpenses"`
	Operatingincome                                        int64     `json:"operatingIncome"`
	Tradeandnontradepayables                               int64     `json:"tradeAndNonTradePayables"`
	Payoutratio                                            float64   `json:"payoutRatio"`
	Pricetobookvalue                                       float64   `json:"priceToBookValue"`
	Priceearnings                                          float64   `json:"priceEarnings"`
	Pricetoearningsratio                                   float64   `json:"priceToEarningsRatio"`
	Propertyplantequipmentnet                              int64     `json:"propertyPlantEquipmentNet"`
	Preferreddividendsincomestatementimpact                int       `json:"preferredDividendsIncomeStatementImpact"`
	Sharepriceadjustedclose                                float64   `json:"sharePriceAdjustedClose"`
	Pricesales                                             float64   `json:"priceSales"`
	Pricetosalesratio                                      float64   `json:"priceToSalesRatio"`
	Tradeandnontradereceivables                            int64     `json:"tradeAndNonTradeReceivables"`
	Accumulatedretainedearningsdeficit                     int64     `json:"accumulatedRetainedEarningsDeficit"`
	Revenues                                               int64     `json:"revenues"`
	Revenuesusd                                            int64     `json:"revenuesUSD"`
	Researchanddevelopmentexpense                          int64     `json:"researchAndDevelopmentExpense"`
	Sharebasedcompensation                                 int       `json:"shareBasedCompensation"`
	Sellinggeneralandadministrativeexpense                 int64     `json:"sellingGeneralAndAdministrativeExpense"`
	Sharefactor                                            int       `json:"shareFactor"`
	Shares                                                 int64     `json:"shares"`
	Weightedaverageshares                                  int64     `json:"weightedAverageShares"`
	Weightedaveragesharesdiluted                           int64     `json:"weightedAverageSharesDiluted"`
	Salespershare                                          float64   `json:"salesPerShare"`
	Tangibleassetvalue                                     int64     `json:"tangibleAssetValue"`
	Taxassets                                              int       `json:"taxAssets"`
	Incometaxexpense                                       int       `json:"incomeTaxExpense"`
	Taxliabilities                                         int       `json:"taxLiabilities"`
	Tangibleassetsbookvaluepershare                        float64   `json:"tangibleAssetsBookValuePerShare"`
	Workingcapital                                         int64     `json:"workingCapital"`
	Assetsaverage                                          int64     `json:"assetsAverage,omitempty"`
	Assetturnover                                          float64   `json:"assetTurnover,omitempty"`
	Averageequity                                          int64     `json:"averageEquity,omitempty"`
	Investedcapitalaverage                                 int64     `json:"investedCapitalAverage,omitempty"`
	Returnonaverageassets                                  float64   `json:"returnOnAverageAssets,omitempty"`
	Returnonaverageequity                                  float64   `json:"returnOnAverageEquity,omitempty"`
	Returnoninvestedcapital                                float64   `json:"returnOnInvestedCapital,omitempty"`
	Returnonsales                                          float64   `json:"returnOnSales,omitempty"`
}

// MarketHolidays
type MarketHolidays struct {
	Exchange string    `json:"exchange"`
	Name     string    `json:"name"`
	Date     time.Time `json:"date"`
	Status   string    `json:"status"`
	Open     time.Time `json:"open,omitempty"`
	Close    time.Time `json:"close,omitempty"`
}

// MarketStatusResponse []
type MarketStatusResponse []struct {
	Market     string    `json:"market"`
	Servertime time.Time `json:"serverTime"`
	Exchanges  struct {
		Nyse   string `json:"nyse"`
		Nasdaq string `json:"nasdaq"`
		Otc    string `json:"otc"`
	} `json:"exchanges"`
	Currencies struct {
		Fx     string `json:"fx"`
		Crypto string `json:"crypto"`
	} `json:"currencies"`
}

// MarketStatus
type MarketStatus struct {
	InsertDate time.Time `json:"insert_date"`
	Market     string    `json:"market"`
	Servertime time.Time `json:"serverTime"`
	Nyse       string    `json:"nyse"`
	Nasdaq     string    `json:"nasdaq"`
	Otc        string    `json:"otc"`
	Fx         string    `json:"fx"`
	Crypto     string    `json:"crypto"`
}

// StockExchanges []
type StockExchanges struct {
	ID     int    `json:"id"`
	Type   string `json:"type"`
	Market string `json:"market"`
	Mic    string `json:"mic"`
	Name   string `json:"name"`
	Tape   string `json:"tape"`
	Code   string `json:"code,omitempty"`
}

// ConditionMappings
type ConditionsMapping struct {
	Num0  string `json:"0"`
	Num1  string `json:"1"`
	Num2  string `json:"2"`
	Num3  string `json:"3"`
	Num4  string `json:"4"`
	Num5  string `json:"5"`
	Num6  string `json:"6"`
	Num7  string `json:"7"`
	Num8  string `json:"8"`
	Num9  string `json:"9"`
	Num10 string `json:"10"`
	Num11 string `json:"11"`
	Num12 string `json:"12"`
	Num13 string `json:"13"`
	Num14 string `json:"14"`
	Num15 string `json:"15"`
	Num16 string `json:"16"`
	Num17 string `json:"17"`
	Num18 string `json:"18"`
	Num19 string `json:"19"`
	Num20 string `json:"20"`
	Num21 string `json:"21"`
	Num22 string `json:"22"`
	Num23 string `json:"23"`
	Num24 string `json:"24"`
	Num25 string `json:"25"`
	Num26 string `json:"26"`
	Num27 string `json:"27"`
	Num28 string `json:"28"`
	Num29 string `json:"29"`
	Num30 string `json:"30"`
	Num31 string `json:"31"`
	Num32 string `json:"32"`
	Num33 string `json:"33"`
	Num34 string `json:"34"`
	Num35 string `json:"35"`
	Num36 string `json:"36"`
	Num37 string `json:"37"`
	Num38 string `json:"38"`
	Num39 string `json:"39"`
	Num40 string `json:"40"`
	Num41 string `json:"41"`
	Num42 string `json:"42"`
	Num43 string `json:"43"`
	Num44 string `json:"44"`
	Num45 string `json:"45"`
	Num46 string `json:"46"`
	Num47 string `json:"47"`
	Num48 string `json:"48"`
	Num49 string `json:"49"`
	Num50 string `json:"50"`
	Num51 string `json:"51"`
	Num52 string `json:"52"`
	Num53 string `json:"53"`
	Num54 string `json:"54"`
}

// CryptoExchangesResponse []
type CryptoExchangesResponse []struct {
	ID     int    `json:"id"`
	Market string `json:"market"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Tier   string `json:"tier"`
	Locale string `json:"locale"`
}

// CryptoExchanges
type CryptoExchanges struct {
	InsertDate time.Time `json:"insert_date"`
	ID         int       `json:"id"`
	Market     string    `json:"market"`
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	Tier       string    `json:"tier"`
	Locale     string    `json:"locale"`
}

// DailyOpenClose
type DailyOpenClose struct {
	Status     string  `json:"status"`
	From       string  `json:"from"`
	Symbol     string  `json:"symbol"`
	Open       int     `json:"open"`
	High       float64 `json:"high"`
	Low        float64 `json:"low"`
	Close      float64 `json:"close"`
	Volume     int     `json:"volume"`
	AfterHours float64 `json:"afterHours"`
	PreMarket  float64 `json:"preMarket"`
}

// AggregatesBarsResults
type (
	AggregatesBarsResults struct {
		V  float64 `json:"v"`
		Vw float64 `json:"vw"`
		O  float64 `json:"o"`
		C  float64 `json:"c"`
		H  float64 `json:"h"`
		L  float64 `json:"l"`
		T  float64 `json:"t"`
		N  int     `json:"n"`
	}
	AggregatesBarsResponse struct {
		Ticker       string                  `json:"ticker"`
		Status       string                  `json:"status"`
		Querycount   int                     `json:"queryCount"`
		Resultscount int                     `json:"resultsCount"`
		Adjusted     bool                    `json:"adjusted"`
		Results      []AggregatesBarsResults `json:"results"`
		RequestID    string                  `json:"request_id"`
		Count        int                     `json:"count"`
	}
	AggregatesBars struct {
		InsertDate   time.Time `json:"insert_date"`
		Ticker       string    `json:"ticker"`
		Status       string    `json:"status"`
		Querycount   int       `json:"queryCount"`
		Resultscount int       `json:"resultsCount"`
		Adjusted     bool      `json:"adjusted"`
		V            float64   `json:"v"`
		Vw           float64   `json:"vw"`
		O            float64   `json:"o"`
		C            float64   `json:"c"`
		H            float64   `json:"h"`
		L            float64   `json:"l"`
		T            time.Time `json:"t"`
		N            int       `json:"n"`
		RequestID    string    `json:"request_id"`
		Multiplier   int       `json:"multiplier"`
		Timespan     string    `json:"timespan"`
	}
)

// NewConsumerStruct Used by Consume
type NewConsumerStruct struct {
	AggBarsResponse AggregatesBarsResponse
	Timespan        string
	Multiplier      int
	DBConn          *pg.Conn
}

// NewBatchConsumerStruct Used by Consume
type NewBatchConsumerStruct struct {
	AggBarsResponse []AggregatesBarsResponse
	Timespan        string
	Multiplier      int
	DBConn          *pg.Conn
}

// TickersVxFlattenPayloadBeforeInsert Function that flattens result from '/v3/reference/tickers'
func TickersVxFlattenPayloadBeforeInsert(target TickersVxResponse) []TickerVx {
	var output []TickerVx
	var results = target.Results

	for i := range results {
		var res = results[i]
		r := TickerVx{
			InsertDatetime:  time.Now(),
			Ticker:          res.Ticker,
			Name:            res.Name,
			Market:          res.Market,
			Locale:          res.Locale,
			PrimaryExchange: res.PrimaryExchange,
			Type:            res.Type,
			Active:          res.Active,
			CurrencyName:    res.CurrencyName,
			Cik:             res.Cik,
			CompositeFigi:   res.CompositeFigi,
			ShareClassFigi:  res.ShareClassFigi,
			LastUpdatedUtc:  res.LastUpdatedUtc,
		}
		output = append(output, r)
	}

	return output
}

// TickersFlattenPayloadBeforeInsert Function that flattens result from '/v2/reference/tickers' (depreciated)
func TickersFlattenPayloadBeforeInsert(target TickersResponse) []Tickers {
	var output []Tickers
	tickersInner := target.Tickers // creates TickersInner

	for i := range tickersInner {
		t := tickersInner[i]

		r := Tickers{
			InsertDate:  time.Now(),
			Page:        target.Page,
			Perpage:     target.Perpage,
			Count:       target.Count,
			Status:      target.Status,
			Ticker:      t.Ticker,
			Name:        t.Name,
			Market:      t.Market,
			Locale:      t.Locale,
			Currency:    t.Currency,
			Active:      t.Active,
			Primaryexch: t.Primaryexch,
		}

		if t.Type != nil {
			r.Type = t.Type
		}

		if t.Codes != nil {
			r.Cik = t.Codes.Cik
			r.Figiuid = t.Codes.Figiuid
			r.Scfigi = t.Codes.Scfigi
			r.Cfigi = t.Codes.Cfigi
			r.Figi = t.Codes.Figi
		}

		r.Updated = t.Updated
		r.URL = t.URL

		if t.Attrs != nil {
			r.Currencyname = t.Attrs.Currencyname
			r.Currency = t.Attrs.Currency
			r.Basename = t.Attrs.Basename
			r.Base = t.Attrs.Base
		}
		output = append(output, r)
	}
	return output
}

// TickerTypesFlattenPayloadBeforeInsert Function that flattens result from '/v2/reference/types'
func TickerTypesFlattenPayloadBeforeInsert(target *TickerTypeResponse) TickerType {
	var Tt TickerType
	Tt.Cs = target.Results.Types.Cs
	Tt.Adr = target.Results.Types.Adr
	Tt.Nvdr = target.Results.Types.Nvdr
	Tt.Gdr = target.Results.Types.Gdr
	Tt.Sdr = target.Results.Types.Sdr
	Tt.Cef = target.Results.Types.Cef
	Tt.Etp = target.Results.Types.Etp
	Tt.Reit = target.Results.Types.Reit
	Tt.Mlp = target.Results.Types.Mlp
	Tt.Wrt = target.Results.Types.Wrt
	Tt.Pub = target.Results.Types.Pub
	Tt.Nyrs = target.Results.Types.Nyrs
	Tt.Unit = target.Results.Types.Unit
	Tt.Right = target.Results.Types.Right
	Tt.Track = target.Results.Types.Track
	Tt.Ltdp = target.Results.Types.Ltdp
	Tt.Rylt = target.Results.Types.Rylt
	Tt.Mf = target.Results.Types.Mf
	Tt.Pfd = target.Results.Types.Pfd
	Tt.Fdr = target.Results.Types.Fdr
	Tt.Ost = target.Results.Types.Ost
	Tt.Fund = target.Results.Types.Fund
	Tt.Sp = target.Results.Types.Sp
	Tt.Si = target.Results.Types.Si
	Tt.Index = target.Results.IndexTypes.Index
	Tt.Etf = target.Results.IndexTypes.Etf
	Tt.Etn = target.Results.IndexTypes.Etf
	Tt.Etmf = target.Results.IndexTypes.Etmf
	Tt.Settlement = target.Results.IndexTypes.Settlement
	Tt.Spot = target.Results.IndexTypes.Spot
	Tt.Subprod = target.Results.IndexTypes.Subprod
	Tt.Wc = target.Results.IndexTypes.Wc
	Tt.Alphaindex = target.Results.IndexTypes.Alphaindex
	return Tt
}

// msToTime Function that takes in the string within the json, and returns the time.Unix element.
func msToTime(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, msInt*int64(time.Millisecond)), nil
}

// AggBarFlattenPayloadBeforeInsert Function that flattens result from '/v2/aggs/ticker/{stocksTicker}/range/{multiplier}/{timespan}/{from}/{to}'
func AggBarFlattenPayloadBeforeInsert(target AggregatesBarsResponse, timespan string, multiplier int) []AggregatesBars {
	var output []AggregatesBars
	results := target.Results

	for i := range results {
		// convert the string time into time.Unix time
		var r = results[i]
		t, err := msToTime(strconv.FormatInt(int64(r.T), 10))
		if err != nil {
			fmt.Println("Some Error parsing date: ", err)
		}

		// flatten here
		newArr := AggregatesBars{
			InsertDate:   time.Now(),
			Ticker:       target.Ticker,
			Status:       target.Status,
			Querycount:   target.Querycount,
			Resultscount: target.Resultscount,
			Adjusted:     target.Adjusted,
			V:            r.V,
			Vw:           r.Vw,
			O:            r.O,
			C:            r.C,
			H:            r.H,
			L:            r.L,
			T:            t,
			N:            r.N,
			RequestID:    target.RequestID,
			Multiplier:   multiplier,
			Timespan:     timespan,
		}
		output = append(output, newArr)
	}
	return output
}

func TickerNews2FlattenPayloadBeforeInsert(target TickerNews2Response) []TickerNews2 {
	var output []TickerNews2
	results := target.Results
	for i := range results {
		var r = results[i]
		newArr := TickerNews2{
			ID:                   r.ID,
			PublisherName:        r.Publisher.Name,
			PublisherHomepageURL: r.Publisher.HomepageURL,
			PublisherLogoURL:     r.Publisher.LogoURL,
			PublisherFaviconURL:  r.Publisher.FaviconURL,
			Title:                r.Title,
			Author:               r.Author,
			PublishedUtc:         r.PublishedUtc,
			ArticleURL:           r.ArticleURL,
			Tickers:              r.Tickers,
			AmpURL:               r.AmpURL,
			ImageURL:             r.ImageURL,
			Description:          r.Description,
			Keywords:             r.Keywords,
			Status:               target.Status,
			RequestID:            target.RequestID,
			Count:                target.Count,
			NextURL:              target.NextURL,
		}
		output = append(output, newArr)
	}
	return output
}

func (aggsConnCombo *NewConsumerStruct) Consume(batch rmq.Deliveries) {
	conn := aggsConnCombo.DBConn
	aggBarsResponse := aggsConnCombo.AggBarsResponse
	payloads := batch.Payloads()

	for _, payload := range payloads {

		if err := json.Unmarshal([]byte(payload), &aggBarsResponse); err != nil {
			fmt.Println("Something Json Error")
			if err := batch.Reject(); err != nil {
				fmt.Println("Something Reject Error")
			}
		}

		aggs := AggBarFlattenPayloadBeforeInsert(aggBarsResponse, aggsConnCombo.Timespan, aggsConnCombo.Multiplier)
		if len(aggs) > 0 {
			_, err := conn.Model(&aggs).OnConflict("(t, vw, multiplier, timespan, ticker, o, h, l, c) DO NOTHING").Insert()
			if err != nil {
				panic(err)
			}

		}

		if errors := batch.Ack(); len(errors) > 0 {
			e := fmt.Sprintf("Something Ack Error:: %s\n", errors)
			fmt.Printf(e)
		}
	}
}

// Grouped Daily Bars
type (
	GroupedDailyResults struct {
		Tcap string  `json:"T"`
		V    int     `json:"v"`
		O    float64 `json:"o"`
		C    float64 `json:"c"`
		H    float64 `json:"h"`
		L    float64 `json:"l"`
		T    int64   `json:"t"`
	}
	GroupedDailyBarsResponse struct {
		Status       string                `json:"status"`
		Querycount   int                   `json:"queryCount"`
		Resultscount int                   `json:"resultsCount"`
		Adjusted     bool                  `json:"adjusted"`
		Results      []GroupedDailyResults `json:"results"`
	}
	GroupedDailyBars struct {
		InsertDate   time.Time `json:"insert_date"`
		Status       string    `json:"status"`
		Querycount   int       `json:"queryCount"`
		Resultscount int       `json:"resultsCount"`
		Adjusted     bool      `json:"adjusted"`
		Tcap         string    `json:"T"`
		V            int       `json:"v"`
		O            float64   `json:"o"`
		C            float64   `json:"c"`
		H            float64   `json:"h"`
		L            float64   `json:"l"`
		T            int64     `json:"t"`
	}
)

// Previous Close
type (
	PreviousCloseResultsResponse struct {
		Tcap string  `json:"T"`
		V    int     `json:"v"`
		Vw   float64 `json:"vw"`
		O    float64 `json:"o"`
		C    float64 `json:"c"`
		H    float64 `json:"h"`
		L    float64 `json:"l"`
		T    int64   `json:"t"`
		N    int     `json:"n"`
	}
	PreviousCloseResponse struct {
		Ticker       string                         `json:"ticker"`
		Querycount   int                            `json:"queryCount"`
		Resultscount int                            `json:"resultsCount"`
		Adjusted     bool                           `json:"adjusted"`
		Results      []PreviousCloseResultsResponse `json:"results"`
		Status       string                         `json:"status"`
		RequestID    string                         `json:"request_id"`
		Count        int                            `json:"count"`
	}
	PreviousClose struct {
		InsertDate   time.Time `json:"insert_date"`
		Ticker       string    `json:"ticker"`
		Querycount   int       `json:"queryCount"`
		Resultscount int       `json:"resultsCount"`
		Adjusted     bool      `json:"adjusted"`
		Tcap         string    `json:"T"`
		V            int       `json:"v"`
		Vw           float64   `json:"vw"`
		O            float64   `json:"o"`
		C            float64   `json:"c"`
		H            float64   `json:"h"`
		L            float64   `json:"l"`
		T            int64     `json:"t"`
		N            int       `json:"n"`
		Status       string    `json:"status"`
		RequestID    string    `json:"request_id"`
		Count        int       `json:"count"`
	}
)

// snapshot - all tickers
type (
	Day struct {
		C  float64 `json:"c"`
		H  float64 `json:"h"`
		L  float64 `json:"l"`
		O  float64 `json:"o"`
		V  int     `json:"v"`
		Vw float64 `json:"vw"`
	}
	Lastquote struct {
		Pcap float64 `json:"P"`
		Scap int     `json:"S"`
		P    float64 `json:"p"`
		S    int     `json:"s"`
		T    int64   `json:"t"`
	}
	Lasttrade struct {
		C []int   `json:"c"`
		I string  `json:"i"`
		P float64 `json:"p"`
		S int     `json:"s"`
		T int64   `json:"t"`
		X int     `json:"x"`
	}
	Min struct {
		Av int     `json:"av"`
		C  float64 `json:"c"`
		H  float64 `json:"h"`
		L  float64 `json:"l"`
		O  float64 `json:"o"`
		V  int     `json:"v"`
		Vw float64 `json:"vw"`
	}
	Prevday struct {
		C  float64 `json:"c"`
		H  int     `json:"h"`
		L  float64 `json:"l"`
		O  float64 `json:"o"`
		V  int     `json:"v"`
		Vw float64 `json:"vw"`
	}
	SnapshotTickers struct {
		Day              Day       `json:"day"`
		Lastquote        Lastquote `json:"lastQuote"`
		Lasttrade        Lasttrade `json:"lastTrade"`
		Min              Min       `json:"min"`
		Prevday          Prevday   `json:"prevDay"`
		Ticker           string    `json:"ticker"`
		Todayschange     float64   `json:"todaysChange"`
		Todayschangeperc float64   `json:"todaysChangePerc"`
		Updated          int64     `json:"updated"`
	}
	SnapshotAllTickersResponse struct {
		Status  string            `json:"status"`
		Count   int               `json:"count"`
		Tickers []SnapshotTickers `json:"tickers"`
	}
	SnapshotAllTickers struct {
		InsertDate       time.Time `json:"insert_date"`
		Status           string    `json:"status"`
		Count            int       `json:"count"`
		Day              Day       `json:"day"`
		LastquotePcap    float64   `json:"lastquote_P"`
		LastquoteScap    int       `json:"lastquote_S"`
		LastquoteP       float64   `json:"lastquote_p"`
		LastquoteS       int       `json:"lastquote_s"`
		LastquoteT       int64     `json:"lastquote_t"`
		LasttradeC       []int     `json:"lasttrade_c"`
		LasttradeI       string    `json:"lasttrade_i"`
		LasttradeP       float64   `json:"lasttrade_p"`
		LasttradeS       int       `json:"lasttrade_s"`
		LasttradeT       int64     `json:"lasttrade_t"`
		LasttradeX       int       `json:"lasttrade_x"`
		MinAv            int       `json:"Min_av"`
		MinC             float64   `json:"Min_c"`
		MinH             float64   `json:"Min_h"`
		MinL             float64   `json:"Min_l"`
		MinO             float64   `json:"Min_o"`
		MinV             int       `json:"Min_v"`
		MinVw            float64   `json:"Min_vw"`
		PrevdayC         float64   `json:"prevday_c"`
		PrevdayH         int       `json:"prevday_h"`
		PrevdayL         float64   `json:"prevday_l"`
		PrevdayO         float64   `json:"prevday_o"`
		PrevdayV         int       `json:"prevday_v"`
		PrevdayVw        float64   `json:"prevday_vw"`
		Ticker           string    `json:"ticker"`
		Todayschange     float64   `json:"todaysChange"`
		Todayschangeperc float64   `json:"todaysChangePerc"`
		Updated          int64     `json:"updated"`
	}
)

// snapshot one ticker
type (
	SnapshotOneTickerResults struct {
		Day              Day       `json:"day"`
		Lastquote        Lastquote `json:"lastQuote"`
		Lasttrade        Lasttrade `json:"lastTrade"`
		Min              Min       `json:"min"`
		Prevday          Prevday   `json:"prevDay"`
		Ticker           string    `json:"ticker"`
		Todayschange     float64   `json:"todaysChange"`
		Todayschangeperc float64   `json:"todaysChangePerc"`
		Updated          int64     `json:"updated"`
	}
	SnapshotOneTickerResponse struct {
		Status string                   `json:"status"`
		Ticker SnapshotOneTickerResults `json:"ticker"`
	}
	SnapshotOneTicker struct {
		InsertDate       time.Time `json:"insert_date"`
		Status           string    `json:"status"`
		Day              Day       `json:"day"`
		Lastquote        Lastquote `json:"lastQuote"`
		Lasttrade        Lasttrade `json:"lastTrade"`
		Min              Min       `json:"min"`
		Prevday          Prevday   `json:"prevDay"`
		Ticker           string    `json:"ticker"`
		Todayschange     float64   `json:"todaysChange"`
		Todayschangeperc float64   `json:"todaysChangePerc"`
		Updated          int64     `json:"updated"`
	}
)

// snapshot gainers and losers ticker
type (
	SnapshotGainersLosersTickers struct {
		Day              Day       `json:"day"`
		Lastquote        Lastquote `json:"lastQuote"`
		Lasttrade        Lasttrade `json:"lastTrade"`
		Min              Min       `json:"min"`
		Prevday          Prevday   `json:"prevDay"`
		Ticker           string    `json:"ticker"`
		Todayschange     float64   `json:"todaysChange"`
		Todayschangeperc float64   `json:"todaysChangePerc"`
		Updated          int64     `json:"updated"`
	}
	SnapshotGainersLosersResponse struct {
		Status  string                         `json:"status"`
		Tickers []SnapshotGainersLosersTickers `json:"tickers"`
	}
	SnapshotGainersLosers struct {
		InsertDate       time.Time `json:"insert_date"`
		Status           string    `json:"status"`
		Day              Day       `json:"day"`
		Lastquote        Lastquote `json:"lastQuote"`
		Lasttrade        Lasttrade `json:"lastTrade"`
		Min              Min       `json:"min"`
		Prevday          Prevday   `json:"prevDay"`
		Ticker           string    `json:"ticker"`
		Todayschange     float64   `json:"todaysChange"`
		Todayschangeperc float64   `json:"todaysChangePerc"`
		Updated          int64     `json:"updated"`
	}
)

type (
	TickerNews2Publisher struct {
		Name        string `json:"name"`
		HomepageURL string `json:"homepage_url"`
		LogoURL     string `json:"logo_url"`
		FaviconURL  string `json:"favicon_url"`
	}
	TickerNews2Results struct {
		ID           string               `json:"id"`
		Publisher    TickerNews2Publisher `json:"publisher"`
		Title        string               `json:"title"`
		Author       string               `json:"author"`
		PublishedUtc time.Time            `json:"published_utc"`
		ArticleURL   string               `json:"article_url"`
		Tickers      []string             `json:"tickers"`
		AmpURL       string               `json:"amp_url"`
		ImageURL     string               `json:"image_url"`
		Description  string               `json:"description"`
		Keywords     []string             `json:"keywords"`
	}
	// TickerNews2Response table
	TickerNews2Response struct {
		Results   []TickerNews2Results `json:"results"`
		Status    string               `json:"status"`
		RequestID string               `json:"request_id"`
		Count     int                  `json:"count"`
		NextURL   string               `json:"next_url"`
	}
)

type TickerNews2 struct {
	ID                   string    `json:"id"`
	PublisherName        string    `json:"publisher_name"`
	PublisherHomepageURL string    `json:"publisher_homepage_url"`
	PublisherLogoURL     string    `json:"publisher_logo_url"`
	PublisherFaviconURL  string    `json:"publisher_favicon_url"`
	Title                string    `json:"title"`
	Author               string    `json:"author"`
	PublishedUtc         time.Time `json:"published_utc"`
	ArticleURL           string    `json:"article_url"`
	Tickers              []string  `json:"tickers"`
	AmpURL               string    `json:"amp_url"`
	ImageURL             string    `json:"image_url"`
	Description          string    `json:"description"`
	Keywords             []string  `json:"keywords"`
	Status               string    `json:"status"`
	RequestID            string    `json:"request_id"`
	Count                int       `json:"count"`
	NextURL              string    `json:"next_url"`
}
