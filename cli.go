package teak

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/urfave/cli.v1"
)

//askSecret - asks password from user, does not echo charectors
func askSecret() (secret string, err error) {
	var pbyte []byte
	pbyte, err = terminal.ReadPassword(int(syscall.Stdin))
	if err == nil {
		secret = string(pbyte)
		fmt.Println()
	}
	return secret, err
}

//AskPassword - asks password, prints the given name before asking
func AskPassword(name string) (secret string) {
	fmt.Print(name + ": ")
	secret, _ = askSecret()
	return secret
}

//ArgGetter - this struct and its method are helpers to combine getting args
//from commandline arguments or from reading from console. Also handles errors
//when required arguments are not provided
type ArgGetter struct {
	ctx *cli.Context
	Err error
}

//NewArgGetter - creates a new argument retriever with given context
func NewArgGetter(ctx *cli.Context) (argtr *ArgGetter) {
	argtr = &ArgGetter{
		ctx: ctx,
	}
	return argtr
}

//readInput - reads stdin
func readInput(text *string) (err error) {
	scanner := bufio.NewScanner(os.Stdin)
	*text = ""
	for scanner.Scan() {
		*text = scanner.Text()
		break
	}
	err = scanner.Err()
	return err
}

//GetRequiredString - gives a string argument either from commandline or from
//blocking user input, this method sets the error if required arg-val is empty
func (retriever *ArgGetter) GetRequiredString(key string) (val string) {
	if retriever.Err != nil {
		return val
	}
	val = retriever.ctx.String(key)
	if !retriever.ctx.IsSet(key) && len(val) == 0 {
		fmt.Print(key + "*: ")
		err := readInput(&val)
		if err != nil || len(val) == 0 {
			val = ""
			retriever.Err = fmt.Errorf("Required argument %s not provided", key)
		}
	}
	return val
}

//GetRequiredSecret - gives a string argument either from commandline or from
//blocking user input, this method sets the error if required arg-val is empty
func (retriever *ArgGetter) GetRequiredSecret(key string) (val string) {
	if retriever.Err != nil {
		return val
	}
	val = retriever.ctx.String(key)
	if !retriever.ctx.IsSet(key) && len(val) == 0 {
		fmt.Print(key + "*: ")
		var err error
		val, err = askSecret()
		if err != nil || len(val) == 0 {
			val = ""
			retriever.Err = fmt.Errorf("Required argument %s not provided", key)
		}
	}
	return val
}

//GetRequiredInt - gives a Integer argument either from commandline or from
//blocking user input, this method sets the error if required arg-val is empty
func (retriever *ArgGetter) GetRequiredInt(key string) (val int) {
	if retriever.Err != nil {
		return val
	}
	val = retriever.ctx.Int(key)
	if !retriever.ctx.IsSet(key) && val == 0 {
		fmt.Print(key + ": ")
		var strval string
		err := readInput(&strval)
		if err != nil || len(strval) == 0 {
			val = 0
			retriever.Err = fmt.Errorf("Required argument %s not provided", key)
		} else {
			val, err = strconv.Atoi(strval)
			if err != nil {
				retriever.Err = fmt.Errorf("Invalid value for %s given", key)
				val = 0
			}
		}
	}
	return val
}

//GetOptionalString - retrieves string from commandline, if not provided
//it wont ask again from stdin
func (retriever *ArgGetter) GetOptionalString(key string) (val string) {
	if retriever.Err != nil {
		return val
	}
	val = retriever.ctx.String(key)
	return val
}

//GetOptionalInt - retrieves int from commandline, if not provided
//it wont ask again from stdin
func (retriever *ArgGetter) GetOptionalInt(key string) (val int) {
	if retriever.Err != nil {
		return val
	}
	val = retriever.ctx.Int(key)
	return val
}

//GetOptionalSecret - reads 'secret' if provided in commandline, otherwise
//return empty string
func (retriever *ArgGetter) GetOptionalSecret(key string) (val string) {
	if retriever.Err != nil {
		return val
	}
	val = retriever.ctx.String(key)
	return val
}

//GetStringOr - retrieves string from commandline, if not found returns
//given default
func (retriever *ArgGetter) GetStringOr(key, def string) (val string) {
	val = def
	if retriever.Err != nil {
		return val
	}
	if retriever.ctx.IsSet(key) {
		val = retriever.ctx.String(key)
	}
	return val
}

//GetIntOr - retrieves string int commandline, if not found returns
//given default
func (retriever *ArgGetter) GetIntOr(key string, def int) (val int) {
	val = def
	if retriever.Err != nil {
		return val
	}
	if retriever.ctx.IsSet(key) {
		val = retriever.ctx.Int(key)
	}
	return val
}

//GetSecretOr - retrieves password from commandline, if not found returns
//given default
func (retriever *ArgGetter) GetSecretOr(key, def string) (val string) {
	val = def
	if retriever.Err != nil {
		return val
	}
	if retriever.ctx.IsSet(key) {
		val = retriever.ctx.String(key)
	}
	return val
}
