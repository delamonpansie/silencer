package main

import (
	"flag"
	"fmt"
	"net"
	"regexp"
	"time"

	"github.com/go-yaml/yaml"
	"github.com/hpcloud/tail"
	"go.uber.org/zap"

	"github.com/delamonpansie/silencer/config"
	"github.com/delamonpansie/silencer/filter"
	"github.com/delamonpansie/silencer/logger"
)

var log = &logger.Log

////////////////////////////////////////////////////////////////////////////////

type rule struct {
	name     string
	re       []*regexp.Regexp
	duration time.Duration
}

func newRule(name string, src []string, duration time.Duration) rule {
	re := make([]*regexp.Regexp, len(src))
	for i, s := range src {
		re[i] = regexp.MustCompile(s)
	}
	return rule{name, re, duration}
}

var debugRule = flag.String("debug-rule", "", "enable rule matching logs for given rule-name")

func (rule *rule) match(line string) (ip net.IP, err error) {
	if *debugRule != "" {
		fmt.Printf("matching rule %q\n", rule.name)
		defer func() {
			switch {
			case ip != nil:
				fmt.Printf(" success %q\n", ip.String())
			case err != nil:
				fmt.Printf(" failure %s\n", err.Error())
			default:
				fmt.Printf(" no match\n")
			}
		}()
	}

	src := line
	for i, re := range rule.re {
		m := re.FindStringSubmatch(src)
		if *debugRule != "" {
			fmt.Printf(" [%d] %q\n", i, src)
			fmt.Printf("  re %s\n", re)
			fmt.Printf("  => %q\n", m)
		}
		switch len(m) {
		case 0:
			return nil, nil
		case 1:
			continue
		case 2:
			src = m[1]
		default:
			return nil, fmt.Errorf("rule %s, unexpected matches %q in line %q:",
				rule.name, m, line)
		}
	}
	ip = net.ParseIP(src).To4()
	if ip == nil {
		return nil, fmt.Errorf("rule %s, invalid ip %q", rule.name, src)
	}
	return
}

///////////////////////////////////////////////////////////////////////////////

func tailLog(filename string) <-chan *tail.Line {
	config := tail.Config{
		Follow: true,
		ReOpen: true,
		Logger: &logger.StdLogger{SugaredLogger: log.Sugar()},
	}
	t, err := tail.TailFile(filename, config)
	if err != nil {
		log.Fatal("TailFile failed", zap.Error(err))
	}

	return t.Lines
}

func subnetListContains(subnets []net.IPNet, ip net.IP) bool {
	for _, subnet := range subnets {
		if subnet.Contains(ip) {
			return true
		}
	}
	return false
}

func match(line *tail.Line, rules []rule, whitelist []net.IPNet) (net.IP, *rule) {
	for _, rule := range rules {
		ip, err := rule.match(line.Text)
		if err != nil {
			log.Warn("match failed", zap.String("line", line.Text), zap.Error(err))
			continue
		}

		if ip == nil {
			continue
		}

		if subnetListContains(whitelist, ip) {
			log.Info("whitelisted", zap.Any("ip", ip))
			return nil, nil
		}

		return ip, &rule
	}
	return nil, nil
}

func run(log *zap.Logger, blocker filter.Blocker, lines <-chan *tail.Line, rules []rule, whitelist []net.IPNet) {
	for line := range lines {
		if line.Err != nil {
			log.Warn("tail failed", zap.Error(line.Err))
			continue
		}
		log.Debug("tail", zap.String("line", line.Text))

		if ip, rule := match(line, rules, whitelist); ip != nil {
			log.Info("block",
				zap.Any("ip", ip),
				zap.Duration("duration", rule.duration),
				zap.String("rule_name", rule.name),
			)
			blocker.Block(ip, rule.duration, rule.name)
		}
	}
}

// ///////////////////////////////////////////////////////////////////////////////
func main() {
	flag.Parse()
	if *debugRule != "" {
		*logger.LogFile = ""
	}
	logger.Init()
	cfg := config.Load()

	if *debugRule != "" {
		var matchedLogFile []config.LogFile
		for _, logFile := range cfg.LogFile {
			var matchedRule []config.Rule
			for _, rule := range logFile.Rule {
				if rule.Name == *debugRule {
					matchedRule = append(matchedRule, rule)
				}
			}
			if len(matchedRule) > 0 {
				logFile.Rule = matchedRule
				matchedLogFile = append(matchedLogFile, logFile)
			}
		}
		if len(matchedLogFile) == 0 {
			log.Fatal("no matching rules found")
		}
		cfg.LogFile = matchedLogFile

		data, err := yaml.Marshal(&cfg)
		if err != nil {
			log.Fatal("marshal config", zap.Error(err))
		}
		fmt.Printf("configuration:\n%s", string(data))
	}

	var blocker filter.Blocker
	switch {
	case *debugRule != "":
		blocker = filter.NewDummy()
	case cfg.Filter.IPSet != nil:
		blocker = filter.NewIPset(cfg.Filter.IPSet.Set)
	case cfg.Filter.NFT != nil:
		blocker = filter.NewNFT(cfg.Filter.NFT.Table, cfg.Filter.NFT.Set)
	default:
		panic("not reached")
	}

	for _, logFile := range cfg.LogFile {
		rules := make([]rule, len(logFile.Rule))
		for i, ruleConfig := range logFile.Rule {
			rules[i] = newRule(ruleConfig.Name, ruleConfig.Re, ruleConfig.Duration)
		}

		go run(
			log.With(zap.String("filename", logFile.FileName)),
			blocker,
			tailLog(logFile.FileName),
			rules,
			cfg.Whitelist,
		)
	}

	select {}
}
