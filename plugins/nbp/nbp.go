// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// Package nbp implements handling of an NBP (Network Boot Program) using an
// URL, e.g. http://[fe80::abcd:efff:fe12:3456]/my-nbp or tftp://10.0.0.1/my-nbp .
// The NBP information is only added if it is requested by the client.
//
// Note that for DHCPv4, unless the URL is prefixed with a "http", "https" or
// "ftp" scheme, the URL will be split into TFTP server name (option 66)
// and Bootfile name (option 67), so the scheme will be stripped out, and it
// will be treated as a TFTP URL. Anything other than host name and file path
// will be ignored (no port, no query string, etc).
//
// For DHCPv6 OPT_BOOTFILE_URL (option 59) is used, and the value is passed
// unmodified. If the query string is specified and contains a "param" key,
// its value is also passed as OPT_BOOTFILE_PARAM (option 60), so it will be
// duplicated between option 59 and 60.
//
// Example usage:
//
// server6:
//   - plugins:
//   - nbp: http://[2001:db8:a::1]/nbp
//
// server4:
//   - plugins:
//   - nbp: tftp://10.0.0.254/nbp
package nbp

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv6"
	"github.com/insomniacslk/dhcp/iana"
	"github.com/toujourser/coredhcp/handler"
	"github.com/toujourser/coredhcp/logger"
	"github.com/toujourser/coredhcp/plugins"
)

var log = logger.GetLogger("plugins/nbp")

// Plugin wraps plugin registration information
var Plugin = plugins.Plugin{
	Name:   "nbp",
	Setup6: setup6,
	Setup4: setup4,
}

var (
	opt59, opt60 dhcpv6.Option
	opt66, opt67 *dhcpv4.Option
)

func parseArgs(args ...string) (*url.URL, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Exactly one argument must be passed to NBP plugin, got %d", len(args))
	}
	return url.Parse(args[0])
}

func setup6(args ...string) (handler.Handler6, error) {
	u, err := parseArgs(args...)
	if err != nil {
		return nil, err
	}
	opt59 = dhcpv6.OptBootFileURL(u.String())
	params := u.Query().Get("params")
	if params != "" {
		opt60 = &dhcpv6.OptionGeneric{
			OptionCode: dhcpv6.OptionBootfileParam,
			OptionData: []byte(params),
		}
	}
	log.Printf("loaded NBP plugin for DHCPv6.")
	return nbpHandler6, nil
}

// func setup4(args ...string) (handler.Handler4, error) {
// 	u, err := parseArgs(args...)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var otsn, obfn dhcpv4.Option
// 	switch u.Scheme {
// 	case "http", "https", "ftp":
// 		obfn = dhcpv4.OptBootFileName(u.String())
// 	default:
// 		otsn = dhcpv4.OptTFTPServerName(u.Host)
// 		obfn = dhcpv4.OptBootFileName(u.Path)
// 		opt66 = &otsn
// 	}

// 	opt67 = &obfn
// 	log.Printf("loaded NBP plugin for DHCPv4.")
// 	return nbpHandler4, nil
// }

func setup4(args ...string) (handler.Handler4, error) {
	u, err := parseArgs(args...)
	if err != nil {
		return nil, err
	}

	return func(req, resp *dhcpv4.DHCPv4) (*dhcpv4.DHCPv4, bool) {
		// 获取客户端架构
		archs := req.ClientArch()
		archStr := getArchString(archs)

		// 克隆原始 URL 避免污染全局配置
		modifiedURL := *u // 复制结构体
		// 直接替换路径中的 {arch}（无需处理编码）
		modifiedURL.Path = strings.ReplaceAll(modifiedURL.Path, "{arch}", archStr)
		// modifiedURL.Path = strings.ReplaceAll(modifiedURL.Path, "{efi}", "grub.0")

		switch archStr {
		case "x86":
			modifiedURL.Path = strings.ReplaceAll(modifiedURL.Path, "{efi}", "grubx64.efi")
		case "arm64":
			modifiedURL.Path = strings.ReplaceAll(modifiedURL.Path, "{efi}", "grubaa64.efi")
		case "x86legacy":
			modifiedURL.Path = strings.ReplaceAll(modifiedURL.Path, "{efi}", "grub.0")
		default:
			modifiedURL.Path = strings.ReplaceAll(modifiedURL.Path, "{efi}", "unknown.efi")
		}

		// 生成选项
		var tftpOpt *dhcpv4.Option
		var bootOpt *dhcpv4.Option
		switch modifiedURL.Scheme {
		case "http", "https", "ftp":
			opt := dhcpv4.OptBootFileName(modifiedURL.String())
			bootOpt = &opt
		default:
			tftp := dhcpv4.OptTFTPServerName(modifiedURL.Host)
			tftpOpt = &tftp
			boot := dhcpv4.OptBootFileName(modifiedURL.Path)
			bootOpt = &boot
		}

		// 更新选项
		if tftpOpt != nil && req.IsOptionRequested(dhcpv4.OptionTFTPServerName) {
			resp.Options.Update(*tftpOpt)
		}
		if bootOpt != nil && req.IsOptionRequested(dhcpv4.OptionBootfileName) {
			resp.Options.Update(*bootOpt)
		}

		log.Printf("u.Scheme: %s, u.Host: %s, u.Path: %s, Generated boot path for [%s]: [%s]", u.Scheme, u.Host, u.Path, archStr, modifiedURL.String())
		log.Print("----------------------------------------------------")
		return resp, true
	}, nil

}

func nbpHandler6(req, resp dhcpv6.DHCPv6) (dhcpv6.DHCPv6, bool) {
	if opt59 == nil {
		// nothing to do
		return resp, true
	}
	decap, err := req.GetInnerMessage()
	if err != nil {
		log.Errorf("Could not decapsulate request: %v", err)
		// drop the request, this is probably a critical error in the packet.
		return nil, true
	}
	for _, code := range decap.Options.RequestedOptions() {
		if code == dhcpv6.OptionBootfileURL {
			// bootfile URL is requested
			resp.AddOption(opt59)
		} else if code == dhcpv6.OptionBootfileParam {
			// optionally add opt60, bootfile params, if requested
			if opt60 != nil {
				resp.AddOption(opt60)
			}
		}
	}
	log.Debugf("Added NBP %s to request", opt59)
	return resp, true
}

func nbpHandler4(req, resp *dhcpv4.DHCPv4) (*dhcpv4.DHCPv4, bool) {
	if opt67 == nil {
		// nothing to do
		return resp, true
	}
	if req.IsOptionRequested(dhcpv4.OptionTFTPServerName) && opt66 != nil {
		resp.Options.Update(*opt66)
		log.Debugf("Added NBP %s / %s to request", opt66, opt67)
	}
	if req.IsOptionRequested(dhcpv4.OptionBootfileName) {
		resp.Options.Update(*opt67)
		log.Debugf("Added NBP %s to request", opt67)
	}
	return resp, true
}

func getArchString(archs []iana.Arch) string {
	if len(archs) == 0 {
		return "default"
	}
	switch archs[0] { // 以第一个架构类型为准
	case iana.EFI_X86_64:
		return "x86"
	case iana.EFI_ARM64:
		return "arm64"
	case iana.INTEL_X86PC:
		return "x86legacy"
	default:
		log.Warnf("Unknown architecture type: %d", archs[0])
		return "unknown"
	}
}
