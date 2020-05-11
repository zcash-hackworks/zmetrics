package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ybbus/jsonrpc"
	"github.com/zcash-hackworks/zTypes"
)

var cfgFile string
var log *logrus.Entry
var logger = logrus.New()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zmetrics",
	Short: "zmetrics gathers data about the Zcash blockchain",
	Long:  `zmetrics gathers data about the Zcash blockchain`,
	Run: func(cmd *cobra.Command, args []string) {
		opts := &Options{
			RPCUser:     viper.GetString("rpc-user"),
			RPCPassword: viper.GetString("rpc-password"),
			RPCHost:     viper.GetString("rpc-host"),
			RPCPort:     viper.GetString("rpc-port"),
		}
		basicAuth := base64.StdEncoding.EncodeToString([]byte(opts.RPCUser + ":" + opts.RPCPassword))
		rpcClient := jsonrpc.NewClientWithOpts("http://"+opts.RPCHost+":"+opts.RPCPort,
			&jsonrpc.RPCClientOpts{
				CustomHeaders: map[string]string{
					"Authorization": "Basic " + basicAuth,
				}})
		if err := tryWritingFile(viper.GetString("output-dir")); err != nil {
			log.Fatalf("Test to write metrics file failed: %s", err)
		}
		if err := generateMetrics(rpcClient); err != nil {
			log.Fatalf("Failed to write metrics file: %s", err)
		}

	},
}

func generateMetrics(rpcClient jsonrpc.RPCClient) (err error) {
	outputDir := viper.GetString("output-dir")
	numBlocks := viper.GetInt("num-blocks")
	if numBlocks == 0 {
		numBlocks = 10
	}
	var startHeight *int = new(int)
	if viper.GetInt("start-height") == 0 {
		startHeight, err = getCurrentHeight(rpcClient)
		if err != nil {
			return err
		}
	} else {
		*startHeight = viper.GetInt("start-height")
	}
	var endHeight *int = new(int)
	if viper.GetInt("end-height") == 0 {
		*endHeight = *startHeight - numBlocks
	} else {
		*endHeight = viper.GetInt("end-height")
	}

	fmt.Printf("Getting metrics startng at %d through %d\n", *startHeight, *endHeight)
	metrics, err := getBlockRangeMetrics(startHeight, endHeight, rpcClient)
	if err != nil {
		return err
	}

	if viper.GetString("output-format") == "html" {
		return writeMetricsHTML(metrics)
	}
	blockFile := outputDir + "/zcashmetrics.json"
	blockJSON, err := json.MarshalIndent(metrics, "", "    ")
	if err != nil {
		log.Fatalln("generateMetrics MarshalIndent error: %s", err)
	}
	return ioutil.WriteFile(blockFile, blockJSON, 0644)

}

func writeMetricsHTML(metrics []*zTypes.BlockMetric) (err error) {
	outputDir := viper.GetString("output-dir")
	for _, entry := range metrics {
		blockFilePath := outputDir + "/" + strconv.Itoa(entry.Height) + ".html"
		blockFileHandle, err := os.Create(blockFilePath)
		if err != nil {
			return err
		}
		blockHTMLtmpl, err := template.ParseFiles("block.template.html")
		if err != nil {
			return err
		}

		err = blockHTMLtmpl.Execute(blockFileHandle, entry)
		if err != nil {
			return err
		}
		blockFileHandle.Close()
	}
	return nil
}

func getBlockRangeMetrics(startHeight *int, endHeight *int, rpcClient jsonrpc.RPCClient) ([]*zTypes.BlockMetric, error) {
	if startHeight == nil {
		currentHeight, err := getCurrentHeight(rpcClient)
		if err != nil {
			return nil, err
		}
		startHeight = currentHeight
	}
	if endHeight == nil {
		*endHeight = *startHeight - 100
		if *endHeight < 0 {
			*endHeight = 0
		}
	}
	if *endHeight > *startHeight {
		return nil, fmt.Errorf("End height before Start height, bailing")
	}
	var blockMetrics []*zTypes.BlockMetric

	for height := *endHeight; height <= *startHeight; height++ {
		var block *zTypes.Block
		log.Debugf("Calling getblock for block %d", height)
		err := rpcClient.CallFor(&block, "getblock", strconv.Itoa(height), 2)
		if err != nil {
			return nil, err
		}

		//var blockMetric *BlockMetric
		blockMetric := &zTypes.BlockMetric{
			Height:           height,
			SaplingValuePool: block.SaplingValuePool(),
			SproutValuePool:  block.SproutValuePool(),
			Size:             block.Size,
			Time:             block.Time,
		}

		for _, tx := range block.TX {
			blockMetric.NumberofTransactions = blockMetric.NumberofTransactions + 1
			if tx.IsTransparent() {
				blockMetric.NumberofTransparent = blockMetric.NumberofTransparent + 1
			} else if tx.IsMixed() {
				blockMetric.NumberofMixed = blockMetric.NumberofMixed + 1
			} else if tx.IsShielded() {
				blockMetric.NumberofShielded = blockMetric.NumberofShielded + 1
			}
		}

		blockMetrics = append(blockMetrics, blockMetric)
	}
	return blockMetrics, nil
}

func getCurrentHeight(rpcClient jsonrpc.RPCClient) (currentHeight *int, err error) {
	var blockChainInfo *zTypes.GetBlockchainInfo
	if err := rpcClient.CallFor(&blockChainInfo, "getblockchaininfo"); err != nil {
		return nil, err
	}
	height := &blockChainInfo.Blocks
	return height, nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is current directory, zmetrics.yaml)")
	rootCmd.PersistentFlags().Uint32("log-level", uint32(logrus.InfoLevel), "log level (logrus 1-7)")
	rootCmd.PersistentFlags().String("rpc-user", "zcashrpc", "rpc user account")
	rootCmd.PersistentFlags().String("rpc-password", "notsecret", "rpc password")
	rootCmd.PersistentFlags().String("rpc-host", "127.0.0.1", "rpc host")
	rootCmd.PersistentFlags().String("rpc-port", "38232", "rpc port")

	rootCmd.PersistentFlags().Int("start-height", 0, "Starting block height, defaults to current height (working backwards)")
	rootCmd.PersistentFlags().Int("end-height", 0, "Ending block height (working backwards)")
	rootCmd.PersistentFlags().Int("num-blocks", 10, "Number of blocks")
	rootCmd.PersistentFlags().String("output-dir", "./blocks", "Output directory")
	rootCmd.PersistentFlags().String("output-format", "json", "Output format")

	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.SetDefault("log-level", int(logrus.InfoLevel))

	viper.BindPFlag("rpc-user", rootCmd.PersistentFlags().Lookup("rpc-user"))
	viper.SetDefault("rpc-user", "zcashrpc")
	viper.BindPFlag("rpc-password", rootCmd.PersistentFlags().Lookup("rpc-password"))
	viper.SetDefault("rpc-password", "notsecret")
	viper.BindPFlag("rpc-host", rootCmd.PersistentFlags().Lookup("rpc-host"))
	viper.SetDefault("rpc-host", "127.0.0.1")
	viper.BindPFlag("rpc-port", rootCmd.PersistentFlags().Lookup("rpc-port"))
	viper.SetDefault("rpc-port", "38232")

	viper.BindPFlag("start-height", rootCmd.PersistentFlags().Lookup("start-height"))
	viper.BindPFlag("end-height", rootCmd.PersistentFlags().Lookup("end-height"))
	viper.BindPFlag("num-blocks", rootCmd.PersistentFlags().Lookup("num-blocks"))
	viper.BindPFlag("output-dir", rootCmd.PersistentFlags().Lookup("output-dir"))
	viper.BindPFlag("output-format", rootCmd.PersistentFlags().Lookup("output-format"))

	logger.SetFormatter(&logrus.TextFormatter{
		//DisableColors:          true,
		//FullTimestamp:          true,
		//DisableLevelTruncation: true,
	})

	onexit := func() {
		fmt.Printf("zmetric died with a Fatal error. Check logfile for details.\n")
	}

	log = logger.WithFields(logrus.Fields{
		"app": "zmetric",
	})

	log.Logger.SetLevel(7)

	logrus.RegisterExitHandler(onexit)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Look in the current directory for a configuration file
		viper.AddConfigPath(".")
		// Viper auto appends extention to this config name
		// For example, lightwalletd.yml
		viper.SetConfigName("zmetrics")
	}

	// Replace `-` in config options with `_` for ENV keys
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.
	var err error
	if err = viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

}
