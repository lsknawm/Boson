import json
import os

# ================= 配置区域 =================
INPUT_FILE = '../questions.json'
OUTPUT_FILE = '../questions.json'

def ensure_field(node, field, default_value):
    """
    [原子操作]
    1. 如果字段不存在，补全默认值。
    2. 如果字段存在但为 None（且默认值不是None），则保持 None。
    3. 特殊处理：如果字段是 code/image 且值是空字符串 ""，强制转为 None/默认值。
    """
    if field not in node:
        node[field] = default_value
    else:
        # 数据清洗：将空字符串 "" 视作 null
        if field == 'code' and node[field] == "":
            node[field] = None
        elif field == 'image' and node[field] == "":
            node[field] = None

def normalize_node(node, is_content=False):
    """
    原地修改节点，补全字段
    """
    if not isinstance(node, dict):
        return

    # --- 核心字段补全 ---
    ensure_field(node, 'code', None)
    ensure_field(node, 'code_error', False)

    # [新增] 代码运行次数，默认为 0
    ensure_field(node, 'code_run_count', 0)

    ensure_field(node, 'has_image', False)
    ensure_field(node, 'image', None)

    # 确保 text 字段存在
    if 'text' not in node:
        node['text'] = ""

    # 题干特有字段
    if is_content:
        ensure_field(node, 'debug_msg', None)

def process_single_choice(questions):
    count = 0
    for q in questions:
        if q.get("type") != "single_choice":
            continue

        # 1. 修正题干 (Content)
        if "content" not in q or not isinstance(q["content"], dict):
            q["content"] = {}
        normalize_node(q["content"], is_content=True)

        # 2. 修正选项 (Structure -> Options)
        if "structure" not in q or not isinstance(q["structure"], dict):
            q["structure"] = {"layout": "vertical", "options": []}

        options = q["structure"].get("options")
        if options is None or not isinstance(options, list):
            q["structure"]["options"] = []
            print(f"⚠️ 警告: ID {q.get('id')} 的 options 格式错误，已重置。")

        for opt in q["structure"]["options"]:
            normalize_node(opt)

        # 3. 修正解析 (Validation -> Explanation)
        if "validation" not in q or not isinstance(q["validation"], dict):
            q["validation"] = {"answer": "", "explanation": {}}

        if "explanation" not in q["validation"] or not isinstance(q["validation"]["explanation"], dict):
            q["validation"]["explanation"] = {}

        normalize_node(q["validation"]["explanation"])

        count += 1
    return count

def main():
    if not os.path.exists(INPUT_FILE):
        # 如果文件不存在，自动创建一个带 code 的测试数据
        print(f"❌ 找不到文件: {INPUT_FILE}，正在生成测试数据模板...")
        create_dummy_file()
        return

    print(f"📂 正在读取: {INPUT_FILE}...")
    try:
        with open(INPUT_FILE, 'r', encoding='utf-8') as f:
            data = json.load(f)
    except Exception as e:
        print(f"❌ JSON 格式错误: {e}")
        return

    if not isinstance(data, list):
        print("❌ 错误: JSON 根节点必须是数组 []")
        return

    print(f"⚙️ 开始标准化 (新增 code_run_count 字段)...")
    count = process_single_choice(data)

    print(f"💾 保存到: {OUTPUT_FILE}...")
    try:
        with open(OUTPUT_FILE, 'w', encoding='utf-8') as f:
            json.dump(data, f, ensure_ascii=False, indent=2)
        print(f"✅ 完成！处理了 {count} 道单选题。")
    except Exception as e:
        print(f"❌ 保存失败: {e}")

def create_dummy_file():
    # 辅助函数：如果没有文件，生成一个
    dummy_data = [{
        "id": "TEST-NEW-FIELD",
        "type": "single_choice",
        "content": { "text": "测试题目", "code": "print('hello')" },
        "structure": { "options": [{"id":"A", "text":"A"}] }
    }]
    with open(INPUT_FILE, 'w', encoding='utf-8') as f:
        json.dump(dummy_data, f, ensure_ascii=False, indent=2)
    print("✅ 测试文件已生成，请再次运行脚本。")

if __name__ == "__main__":
    main()