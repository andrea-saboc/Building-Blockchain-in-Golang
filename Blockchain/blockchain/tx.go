package blockchain

type TxOutput struct {
	Value  int    //value in tokend
	PubKey string //value that is needed to unlock the token inside value, very complicated scripting language called script
	//cannot dereference the part of the output
}

type TxInput struct {
	//references to the previous outputs
	ID  []byte //references the transaction the output is inside
	Out int    //if the transaction has 3 output and if we want to reference only one of them, at index out
	Sig string //similiar to the pubkey in output
}

// unlock data inside outputs and inputs
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data //signature
}

func (out *TxOutput) CanBeUnlocked(data string) bool { //the data owns the information iside the output
	return out.PubKey == data
}
