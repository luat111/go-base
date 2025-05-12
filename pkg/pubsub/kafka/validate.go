package kafka

import "fmt"

func validateConfigs(conf *Config) error {
	if err := validateRequiredFields(conf); err != nil {
		return err
	}

	return nil
}

func validateRequiredFields(conf *Config) error {
	if len(conf.Brokers) == 0 {
		return errBrokerNotProvided
	}

	if conf.BatchSize <= 0 {
		return fmt.Errorf("batch size must be greater than 0: %w", errBatchSize)
	}

	if conf.BatchBytes <= 0 {
		return fmt.Errorf("batch bytes must be greater than 0: %w", errBatchBytes)
	}

	if conf.BatchTimeout <= 0 {
		return fmt.Errorf("batch timeout must be greater than 0: %w", errBatchTimeout)
	}

	return nil
}
