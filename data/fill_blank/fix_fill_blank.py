import json
import os

INPUT_FILE = '../questions.json'
OUTPUT_FILE = '../questions.json'

def ensure_field(node, field, default_value):
    """原子操作：补全字段，清洗空字符串"""
    if field not in node:
        node[field] = default_value
    else:
        # 清洗空字符串为 None
        if field in ['code', 'image'] and node[field] == "":
            node[field] = None

def normalize_node(node, is_content=False):
    """通用节点清洗：补全 code, image, code_run_count 等"""
    if not isinstance(node, dict):
        return

    # --- 核心字段 ---
    ensure_field(node, 'code', None)
    ensure_field(node, 'code_error', False)
    ensure_field(node, 'code_run_count', 0)  # [核心] 运行次数
    ensure_field(node, 'has_image', False)
    ensure_field(node, 'image', None)

    # --- 文本字段 ---
    if 'text' not in node:
        node['text'] = ""
    # 填空题特有：placeholder (占位符)
    if 'placeholder' not in node and not is_content:
        # 仅针对 blanks 里的节点补全 placeholder
        # 注意：这里不做强制，仅作为防御性检查
        pass

    # --- 题干特有 ---
    if is_content:
        ensure_field(node, 'debug_msg', None)

def process_fill_blank(questions):
    count = 0
    for q in questions:
        # 1. 类型过滤
        if q.get("type") != "fill_blank":
            continue

        # 2. 修正题干 (Content)
        if "content" not in q or not isinstance(q["content"], dict):
            q["content"] = {}
        normalize_node(q["content"], is_content=True)

        # 3. 修正结构 (Structure -> Blanks)
        if "structure" not in q or not isinstance(q["structure"], dict):
            q["structure"] = {"blanks": []}

        # 填空题使用 'blanks' 数组
        blanks = q["structure"].get("blanks")
        if blanks is None or not isinstance(blanks, list):
            q["structure"]["blanks"] = []
        else:
            # 即使是填空位的定义，也为其补全标准字段
            # 这样未来如果需要在填空位显示小图标或代码，前端也能支持
            for blank in blanks:
                normalize_node(blank)

        # 4. 修正解析与答案 (Validation)
        if "validation" not in q or not isinstance(q["validation"], dict):
            q["validation"] = {"answer": {}, "explanation": {}}

        # [特有逻辑] 填空题的 answer 必须是字典 {id: [val1, val2]}
        ans = q["validation"].get("answer")
        if not isinstance(ans, dict):
            print(f"⚠️ ID {q.get('id')} 的 answer 类型错误，重置为空字典 {{}}")
            q["validation"]["answer"] = {}

        # 修正 explanation
        if "explanation" not in q["validation"] or not isinstance(q["validation"]["explanation"], dict):
            q["validation"]["explanation"] = {}
        normalize_node(q["validation"]["explanation"])

        count += 1
    return count

def main():
    if not os.path.exists(INPUT_FILE):
        print(f"❌ 找不到文件: {INPUT_FILE}")
        return

    print(f"📂 读取数据: {INPUT_FILE}...")
    try:
        with open(INPUT_FILE, 'r', encoding='utf-8') as f:
            data = json.load(f)
    except Exception as e:
        print(f"❌ JSON 格式错误: {e}")
        return

    # 兼容单对象
    is_single = False
    if isinstance(data, dict):
        data = [data]
        is_single = True

    print(f"⚙️ 正在修复填空题 (fill_blank)...")
    count = process_fill_blank(data)

    if is_single:
        data = data[0]

    try:
        with open(OUTPUT_FILE, 'w', encoding='utf-8') as f:
            json.dump(data, f, ensure_ascii=False, indent=2)
        print(f"✅ 完成！修正了 {count} 道填空题。")
    except Exception as e:
        print(f"❌ 保存失败: {e}")

if __name__ == "__main__":
    main()