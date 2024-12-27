package tester

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

const PackerCtxKey = "PACKER_CTX_KEY"

type PackerCtx struct {
	ScwClient *scw.Client
}

func NewContext(ctx context.Context) (context.Context, error) {
	cfg, err := scw.LoadConfig()
	if err != nil {
		return nil, err
	}
	activeProfile, err := cfg.GetActiveProfile()
	if err != nil {
		return nil, err
	}

	profile := scw.MergeProfiles(activeProfile, scw.LoadEnvProfile())
	client, err := scw.NewClient(scw.WithProfile(profile))
	if err != nil {
		return nil, err
	}
	return context.WithValue(ctx, PackerCtxKey, &PackerCtx{client}), nil
}

func ExtractCtx(ctx context.Context) *PackerCtx {
	return ctx.Value(PackerCtxKey).(*PackerCtx)
}

func Run(ctx context.Context, packerChecks ...PackerCheck) {
	log.Println("Running tests...")
	ctx, err := NewContext(ctx)
	if err != nil {
		panic(err)
	}

	for i, check := range packerChecks {
		log.Println("Running test", i)
		err := check.Check(ctx)
		if err != nil {
			log.Fatalln(fmt.Sprintf("Packer check %d failed:", i), err)
		}
	}

	os.Exit(0)
}