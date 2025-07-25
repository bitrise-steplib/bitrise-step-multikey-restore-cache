package step

import (
    "errors"
    "fmt"
    "strings"
    "sync"

    "github.com/bitrise-io/go-steputils/v2/cache"
    "github.com/bitrise-io/go-steputils/v2/stepconf"
    "github.com/bitrise-io/go-utils/v2/command"
    "github.com/bitrise-io/go-utils/v2/env"
    "github.com/bitrise-io/go-utils/v2/log"
)

const (
    stepId = "multikey-restore-cache"

    fmtErrNoKeysFound           = "no keys found in input"
    fmtErrFailure               = "save failed"
    fmtErrPartialFailure        = "save failures\n"
    fmtErrPartialFailureDetails = "    - %s\n"
    fmtErrEvaluation            = "keys evaluation failure: %w"
)

type Input struct {
    Verbose        bool   `env:"verbose,required"`
    Keys           string `env:"keys,required"`
    NumFullRetries int    `env:"retries,required"`
}

type MultikeyRestoreCacheStep struct {
    logger         log.Logger
    inputParser    stepconf.InputParser
    commandFactory command.Factory
    envRepo        env.Repository
}

func New(
    logger log.Logger,
    inputParser stepconf.InputParser,
    commandFactory command.Factory,
    envRepo env.Repository,
) MultikeyRestoreCacheStep {
    return MultikeyRestoreCacheStep{
        logger:         logger,
        inputParser:    inputParser,
        commandFactory: commandFactory,
        envRepo:        envRepo,
    }
}

func (step MultikeyRestoreCacheStep) Run() error {

    var input Input
    if err := step.inputParser.Parse(&input); err != nil {
        return err
    }
    stepconf.Print(input)
    step.logger.Println()
    step.logger.EnableDebugLog(input.Verbose)

    keys, evaluationError := input.evaluateKeys()
    if evaluationError != nil {
        return fmt.Errorf(fmtErrEvaluation, evaluationError)
    }

    var wg sync.WaitGroup
    errs := make(chan error, len(keys)) // buffered channel

    for _, keyAndFallbacks := range keys {
        wg.Add(1)

        restore(
            step,
            CacheInput{
                Verbose:         input.Verbose,
                KeyAndFallbacks: keyAndFallbacks,
                NumFullRetries:  input.NumFullRetries,
            },
            &wg,
            errs,
        )
    }

    wg.Wait()
    close(errs)

    if len(errs) > 0 {
        step.logger.Printf(fmtErrPartialFailure)
        for err := range errs {
            step.logger.Printf(fmtErrPartialFailureDetails, err.Error())
        }
    }

    if len(errs) == len(keys) {
        return errors.New(fmtErrFailure)
    }

    return nil
}

func (input Input) evaluateKeys() ([][]string, error) {
    var keys [][]string

    lines := strings.Split(input.Keys, "\n")

    for _, line := range lines {
        keyStrings := strings.Split(line, "||")

        if strings.TrimSpace(keyStrings[0]) == "" && len(keyStrings) == 1 {
            continue
        }

        var alternatives []string
        for _, keyString := range keyStrings {
            key := strings.TrimSpace(keyString)
            if key != "" {
                alternatives = append(alternatives, key)
            }
        }

        keys = append(keys, alternatives)
    }

    if len(keys) == 0 {
        return nil, errors.New(fmtErrNoKeysFound)
    }

    return keys, nil
}

type CacheInput struct {
    Verbose         bool
    KeyAndFallbacks []string
    NumFullRetries  int
}

func restore(
    step MultikeyRestoreCacheStep,
    cacheInput CacheInput,
    wg *sync.WaitGroup,
    errors chan<- error,
) {
    defer wg.Done()

    err := cache.NewRestorer(
        step.envRepo,
        step.logger,
        step.commandFactory,
        nil,
    ).Restore(cache.RestoreCacheInput{
        StepId:         stepId,
        Verbose:        cacheInput.Verbose,
        Keys:           cacheInput.KeyAndFallbacks,
        NumFullRetries: cacheInput.NumFullRetries,
    })

    if err != nil {
        errors <- err
    }
}
