package msgpackzip

import (
	"bytes"
	"fmt"
	"sort"
)

type compressor struct {
	input []byte
}

func newCompressor(b []byte) *compressor {
	return &compressor{input: b}
}

func Compress(input []byte) (output []byte, err error) {
	return newCompressor(input).run()
}

func (c *compressor) run() (output []byte, err error) {

	freqs, err := c.collectFrequencies()
	if err != nil {
		return nil, err
	}
	keys, err := c.sortIntoKeys(freqs)
	if err != nil {
		return nil, err
	}
	output, err = c.output(keys)
	return output, err
}

func (c *compressor) collectFrequencies() (ret map[interface{}]int, err error) {

	ret = make(map[interface{}]int)
	hooks := msgpackDecoderHooks{
		mapKeyHook: func(d decodeStack) (decodeStack, error) {
			d.hooks = msgpackDecoderHooks{
				stringHook: func(l msgpackInt, s string) error {
					ret[s]++
					return nil
				},
				intHook: func(l msgpackInt) error {
					i, err := l.toInt64()
					if err != nil {
						return err
					}
					ret[i]++
					return nil
				},
				fallthroughHook: func(i interface{}, s string) error {
					return fmt.Errorf("bad map key (type %T)", i)
				},
			}
			return d, nil
		},
	}
	err = newMsgpackDecoder(bytes.NewReader(c.input)).run(hooks)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *compressor) sortIntoKeys(freqs map[interface{}]int) (keys map[interface{}]uint, err error) {
	type tuple struct {
		key  interface{}
		freq int
	}
	tuples := make([]tuple, len(freqs))
	var i int
	for k, v := range freqs {
		tuples[i] = tuple{k, v}
		i++
	}
	sort.SliceStable(tuples, func(i, j int) bool { return tuples[i].freq > tuples[j].freq })

	ret := make(map[interface{}]uint, len(freqs))
	for i, tup := range tuples {
		ret[tup.key] = uint(i)
	}
	return ret, nil
}

func (c *compressor) output(keys map[interface{}]uint) (output []byte, err error) {

	version := 1
	data, err := c.outputData(keys)
	if err != nil {
		return nil, err
	}
	compressedKeymap, err := c.outputCompressedKeymap(keys)
	if err != nil {
		return nil, err
	}
	return c.outputFinalProduct(version, data, compressedKeymap)
}

func (c *compressor) outputData(keys map[interface{}]uint) (output []byte, err error) {

	var data outputter

	hooks := data.decoderHooks()
	hooks.mapKeyHook = func(d decodeStack) (decodeStack, error) {
		d.hooks = msgpackDecoderHooks{
			intHook: func(l msgpackInt) error {
				i, err := l.toInt64()
				if err != nil {
					return err
				}
				val, ok := keys[i]
				if !ok {
					return fmt.Errorf("unexpected map key: %v", i)
				}
				return data.outputRawUint(val)
			},
			stringHook: func(l msgpackInt, s string) error {
				val, ok := keys[s]
				if !ok {
					return fmt.Errorf("unexpected map key: %q", s)
				}
				return data.outputRawUint(val)
			},
			fallthroughHook: func(i interface{}, s string) error {
				return fmt.Errorf("bad map key (type %T)", i)
			},
		}
		return d, nil
	}

	err = newMsgpackDecoder(bytes.NewReader(c.input)).run(hooks)
	if err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}

func (c *compressor) outputCompressedKeymap(keys map[interface{}]uint) (output []byte, err error) {

	var keymap outputter

	// Now write out a msgpack dictionary for the keymaps;
	// do it but hand, that's simplest for now, rather than pulling
	// in a new encoder.
	err = keymap.outputMapPrefix(msgpackIntFromUint(uint(len(keys))))
	if err != nil {
		return nil, err
	}
	for k, v := range keys {
		// Note that we reverse the map to make decoding easier
		err = keymap.outputInt(msgpackIntFromUint(v))
		if err != nil {
			return nil, err
		}
		err = keymap.outputStringOrUint(k)
		if err != nil {
			return nil, err
		}
	}
	tmp := keymap.Bytes()

	compressedKeymap, err := gzipCompress(tmp)
	if err != nil {
		return nil, err
	}

	return compressedKeymap, nil
}

func (c *compressor) outputFinalProduct(version int, data []byte, compressedKeymap []byte) (output []byte, err error) {

	var ret outputter
	err = ret.outputArrayPrefix(msgpackIntFromUint(uint(3)))
	if err != nil {
		return nil, err
	}
	err = ret.outputInt(msgpackIntFromUint(uint(version)))
	if err != nil {
		return nil, err
	}
	err = ret.outputBinary(msgpackIntFromUint(uint(len(data))), data)
	if err != nil {
		return nil, err
	}
	err = ret.outputBinary(msgpackIntFromUint(uint(len(compressedKeymap))), compressedKeymap)
	if err != nil {
		return nil, err
	}

	return ret.Bytes(), nil
}