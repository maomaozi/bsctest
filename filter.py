from eth_utils import keccak, to_checksum_address

CREATE_FOURMEME_TOKEN_SELECTOR = "0x519ebb10"
CREATOR_3946 = "0x757eba15a64468e6535532fcf093cef90e226f85"
CREATOR_5546 = "0xc6496e138af13c0026e14ffbd32eae6764eab8b2"
INIT_HASH_3946 = "0x3eb722ec5d79ddc2f52880ea62f1b7e7d95c66d4ae0dfe32f988ca9eca52b359"
INIT_HASH_5546 = "0xa410105265035fe66e02e324fbc5f3d69c891340e6644808247a1cb49e4d5da0"

def predict_token_address(input_data_hex):
    try:
        if len(input_data_hex) < 202:
            return None

        hex_content = input_data_hex[202:]
        total_bytes = len(hex_content) / 2
        words = []

        for i in range(0, len(hex_content), 64):
            chunk = hex_content[i:i + 64]
            if len(chunk) < 64: break
            words.append(int(chunk, 16))

        if len(words) < 10 or words[0] != 32:
            return None

        def is_offset(v):
            return v is not None and v >= 64 and v % 32 == 0 and v < total_bytes

        if not is_offset(words[3]) or not is_offset(words[4]):
            return None

        request_id = words[1]
        salt = words[2]
        token_id = words[9]

        quote_addr = None
        if len(words) > 8:
            quote_addr_hex = Web3.to_hex(words[8])[2:].rjust(40, '0')
            quote_addr = to_checksum_address('0x' + quote_addr_hex)

        creator_type = (token_id >> 10) & 0x3f
        if creator_type == 4:
            creator = CREATOR_5546
            init_hash = INIT_HASH_5546
        else:
            creator = CREATOR_3946
            init_hash = INIT_HASH_3946

        salt_hex = Web3.to_hex(salt)[2:].rjust(64, '0')
        packed_data = (b'\xff' + bytes.fromhex(creator[2:]) + bytes.fromhex(salt_hex) + bytes.fromhex(init_hash[2:]))
        raw_address = keccak(packed_data)[12:]
        token_address = to_checksum_address(raw_address)

        def decode_str(offset_word):
            try:
                start_char_idx = (32 + offset_word) * 2
                if start_char_idx + 64 > len(hex_content): return ""
                len_hex = hex_content[start_char_idx: start_char_idx + 64]
                str_len = int(len_hex, 16)
                content_start = start_char_idx + 64
                content_end = content_start + (str_len * 2)
                if content_end > len(hex_content): return ""
                return bytes.fromhex(hex_content[content_start: content_end]).decode('utf-8', errors='ignore')
            except:
                return None

        return {
            "token": token_address,
            "name": decode_str(words[3]),
            "symbol": decode_str(words[4]),
            "quote": quote_addr
        }
    except Exception:
        return None


if __name__ == "__main__":
    input_data = ""
    if input_data.startswith(CREATE_FOURMEME_TOKEN_SELECTOR):
        prediction = predict_token_address(input_data)