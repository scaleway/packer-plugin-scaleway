// Code generated by "packer-sdc mapstructure-to-hcl2"; DO NOT EDIT.

package scaleway

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// FlatConfigRootVolume is an auto-generated flat version of ConfigRootVolume.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatConfigRootVolume struct {
	Type     *string `mapstructure:"type" cty:"type" hcl:"type"`
	IOPS     *uint32 `mapstructure:"iops" cty:"iops" hcl:"iops"`
	SizeInGB *uint64 `mapstructure:"size_in_gb" cty:"size_in_gb" hcl:"size_in_gb"`
}

// FlatMapstructure returns a new FlatConfigRootVolume.
// FlatConfigRootVolume is an auto-generated flat version of ConfigRootVolume.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*ConfigRootVolume) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatConfigRootVolume)
}

// HCL2Spec returns the hcl spec of a ConfigRootVolume.
// This spec is used by HCL to read the fields of ConfigRootVolume.
// The decoded values from this spec will then be applied to a FlatConfigRootVolume.
func (*FlatConfigRootVolume) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"type":       &hcldec.AttrSpec{Name: "type", Type: cty.String, Required: false},
		"iops":       &hcldec.AttrSpec{Name: "iops", Type: cty.Number, Required: false},
		"size_in_gb": &hcldec.AttrSpec{Name: "size_in_gb", Type: cty.Number, Required: false},
	}
	return s
}
