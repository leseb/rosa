package cluster

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/rosa/pkg/logging"
	"github.com/openshift/rosa/pkg/ocm"
)

var _ = Describe("Validates OCP version", func() {

	const (
		nightly   = "nightly"
		stable    = "stable"
		candidate = "candidate"
		fast      = "fast"
	)
	var client *ocm.Client
	BeforeEach(func() {
		c, err := ocm.NewClient().Logger(logging.NewLogger()).Build()
		Expect(err).NotTo(HaveOccurred())
		client = c
	})

	var _ = Context("when creating a hosted cluster", func() {

		It("OK: Validates successfully a cluster for hosted clusters with a supported version", func() {
			v, err := client.ValidateVersion("4.12.5", []string{"4.12.5"}, stable, false, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(v).To(Equal("openshift-v4.12.5"))
		})

		It("OK: Validates successfully a nightly version of OCP for hosted clusters "+
			"with a supported version", func() {
			v, err := client.ValidateVersion("4.12.0-0.nightly-2023-04-10-222146",
				[]string{"4.12.0-0.nightly-2023-04-10-222146"}, nightly, false, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(v).To(Equal("openshift-v4.12.0-0.nightly-2023-04-10-222146-nightly"))
		})

		It("KO: Fails with a nightly version of OCP for hosted clusters "+
			"in a not supported version", func() {
			v, err := client.ValidateVersion("4.11.0-0.nightly-2022-10-17-040259",
				[]string{"4.11.0-0.nightly-2022-10-17-040259"}, nightly, false, true)
			Expect(err).To(BeEquivalentTo(
				fmt.Errorf("version '4.11.0-0.nightly-2022-10-17-040259' " +
					"is not supported for hosted clusters")))
			Expect(v).To(Equal(""))
		})

		It("OK: Validates successfully the next major release of OCP for hosted clusters "+
			"with a supported version", func() {
			v, err := client.ValidateVersion("4.13.0-rc.2", []string{"4.13.0-rc.2"}, candidate, false, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(v).To(Equal("openshift-v4.13.0-rc.2-candidate"))
		})

		It(`KO: Fails to validate a cluster for a hosted
		cluster when the user provides an unsupported version`, func() {
			v, err := client.ValidateVersion("4.11.5", []string{"4.11.5"}, stable, false, true)
			Expect(err).To(BeEquivalentTo(fmt.Errorf("version '4.11.5' is not supported for hosted clusters")))
			Expect(v).To(BeEmpty())
		})

		It(`KO: Fails to validate a cluster for a hosted cluster
		when the user provides an invalid or malformed version`, func() {
			v, err := client.ValidateVersion("foo.bar", []string{"foo.bar"}, stable, false, true)
			Expect(err).To(BeEquivalentTo(
				fmt.Errorf("version 'foo.bar' was not found")))
			Expect(v).To(BeEmpty())
		})
	})
	var _ = Context("when creating a classic cluster", func() {
		It("OK: Validates successfully a cluster with a supported version", func() {
			v, err := client.ValidateVersion("4.11.0", []string{"4.11.0"}, stable, true, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(v).To(Equal("openshift-v4.11.0"))
		})
	})
})

func TestParseDiskSizeToGigibyte(t *testing.T) {
	type args struct {
		size string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"invalid unit: 1foo", args{"1foo"}, 0, true},
		{"valid unit: 0", args{"0"}, 0, false},
		{"invalid unit no suffix: 1 but return 0", args{"0"}, 0, false},
		{"invalid unit: 1K", args{"1K"}, 0, true},
		{"invalid unit: 1KiB", args{"1KiB"}, 0, true},
		{"invalid unit: 1 MiB", args{"1 MiB"}, 0, true},
		{"invalid unit: 1 mib", args{"1 mib"}, 0, true},
		{"invalid unit: 0 GiB", args{"0 GiB"}, 0, false},
		{"valid unit: 100 G", args{"100 G"}, 93, false},
		{"valid unit: 100GB", args{"100GB"}, 93, false},
		{"valid unit: 100Gb", args{"100Gb"}, 93, false},
		{"valid unit: 100g", args{"100g"}, 93, false},
		{"valid unit: 100GiB", args{"100GiB"}, 100, false},
		{"valid unit: 100gib", args{"100gib"}, 100, false},
		{"valid unit: 100 gib", args{"100 gib"}, 100, false},
		{"valid unit: 100 TB", args{"100 TB"}, 93132, false},
		{"valid unit with spaces: 100 T ", args{"100 T "}, 93132, false},
		{"valid unit: 1000 Ti", args{"1000 Ti"}, 1024000, false},
		{"valid unit: empty string", args{""}, 0, false},
		{"valid unit: -1", args{"-1"}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDiskSizeToGigibyte(tt.args.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDiskSizeToGigibyte() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseDiskSizeToGigibyte() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_machinePoolRooDiskSizeValidator(t *testing.T) {
	type args struct {
		val interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"valid size: 128", args{"128 GiB"}, false},
		{"invalid size: 99", args{"99 GiB"}, true},
		{"invalid size: 65537", args{"65537 GiB"}, true},
		{"invalid size: not a string", args{65537}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := machinePoolRooDiskSizeValidator(tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("machinePoolRooDiskSizeValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
