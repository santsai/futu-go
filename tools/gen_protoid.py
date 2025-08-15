# for generating protoc-gen-go-futu/reqid.go
# $ python gen_protoid.py | gofmt > protoc-gen-go-futu/protoid.go

from futu import ProtoId

def python_dict_to_go_map(py_dict):
    lines_id2name = [f"var protoid_id2name = map[int]string{{"]
    lines_name2id = [f"var protoid_name2id = map[string]int{{"]
    for key, value in py_dict.items():
        key = key.replace("_", "")

        if isinstance(key, str) and isinstance(value, int):
            lines_id2name.append(f'{value}: "{key}",')
            lines_name2id.append(f'"{key}": {value},')

    lines_id2name.append("}")
    lines_name2id.append("}")
    return "\n".join(lines_id2name + lines_name2id)


def gen_all_pushid(py_dict):
    push_ids = py_dict["All_PushId"]
    line = "var protoid_push = []int{"
    line += ",".join([str(x) for x in push_ids])
    line += "}"
    return line


def main():
    items = {k:v for (k,v) in vars(ProtoId).items() if not k.startswith("_")}
    print("package main")
    print(python_dict_to_go_map(items))
    print(gen_all_pushid(items))


if __name__ == "__main__":
    main()
	

