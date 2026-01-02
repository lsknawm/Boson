import json
import os

# ================= 配置区域 =================
INPUT_FILE = '../questions.json'  # 根据实际路径修改
OUTPUT_FILE = '../questions_fixed.json'

def ensure_field(node, field, default_value):
    """
    [原子操作]
    1. 如果字段不存在，补全默认值。
    2. 如果字段存在但为 None（且默认值不是None），则保持 None。
    3. 特殊处理：如果字段是 code/image 且值是空字符串 ""，强制转为 None。
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
    原地修改节点，补全通用的富文本字段 (code, image, text 等)
    """
    if not isinstance(node, dict):
        return

    # --- 核心字段补全 ---
    ensure_field(node, 'code', None)
    ensure_field(node, 'code_error', False)
    ensure_field(node, 'code_run_count', 0) # [新增] 运行次数

    ensure_field(node, 'has_image', False)
    ensure_field(node, 'image', None)

    # 确保 text 字段存在，防止前端报错
    if 'text' not in node:
        node['text'] = ""

    # 题干特有字段 (debug_msg)
    if is_content:
        ensure_field(node, 'debug_msg', None)

def process_single_choice(questions):
    count = 0
    for q in questions:
        # 只处理单选题
        if q.get("type") != "single_choice":
            continue

        # =========================================
        # 1. Top-Level 字段补全 (Subject, UUID等)
        # =========================================

        # [修复] 补全 Subject，默认标记为 Uncategorized，方便后续搜索替换
        ensure_field(q, "subject", "Uncategorized")

        # [建议] 确保有 id (虽然一般都有，但以防万一)
        ensure_field(q, "id", f"UNKNOWN-{count}")

        # =========================================
        # 2. Meta 元数据补全 (Chapter, Difficulty, Score)
        # =========================================
        if "meta" not in q or not isinstance(q["meta"], dict):
            q["meta"] = {}

        ensure_field(q["meta"], "chapter", "General") # 默认章节
        ensure_field(q["meta"], "difficulty", "C")    # 默认难度 C
        ensure_field(q["meta"], "score", 5)           # 默认分值 5

        # =========================================
        # 3. 修正题干 (Content)
        # =========================================
        if "content" not in q or not isinstance(q["content"], dict):
            q["content"] = {}
        normalize_node(q["content"], is_content=True)

        # =========================================
        # 4. 修正结构 (Structure -> Layout & Options)
        # =========================================
        if "structure" not in q or not isinstance(q["structure"], dict):
            q["structure"] = {}

        # [修复] 补全 layout，默认为垂直排列
        ensure_field(q["structure"], "layout", "vertical")

        # 检查 options 列表
        options = q["structure"].get("options")
        if options is None or not isinstance(options, list):
            q["structure"]["options"] = []
            print(f"⚠️ 警告: ID {q.get('id')} 的 options 格式错误，已重置为空列表。")

        # 递归清洗每个选项
        for opt in q["structure"]["options"]:
            normalize_node(opt)

        # =========================================
        # 5. 修正解析 (Validation -> Answer & Explanation)
        # =========================================
        if "validation" not in q or not isinstance(q["validation"], dict):
            q["validation"] = {}

        ensure_field(q["validation"], "answer", "") # 默认空答案

        if "explanation" not in q["validation"] or not isinstance(q["validation"]["explanation"], dict):
            q["validation"]["explanation"] = {}

        normalize_node(q["validation"]["explanation"])

        count += 1
    return count

def main():
    if not os.path.exists(INPUT_FILE):
        print(f"❌ 找不到文件: {INPUT_FILE}，正在生成测试数据...")
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

    print(f"⚙️ 开始标准化 (补全 Subject, Meta, Layout, Code等)...")
    count = process_single_choice(data)

    print(f"💾 保存到: {OUTPUT_FILE}...")
    try:
        with open(OUTPUT_FILE, 'w', encoding='utf-8') as f:
            json.dump(data, f, ensure_ascii=False, indent=2)
        print(f"✅ 完成！成功处理了 {count} 道单选题。")
    except Exception as e:
        print(f"❌ 保存失败: {e}")

def create_dummy_file():
    # 生成一个缺胳膊少腿的测试数据，用来验证脚本是否生效
    dummy_data = [{
        "id": "TEST-MISSING-FIELDS",
        "type": "single_choice",
        # 缺少 subject, meta, structure.layout
        "content": { "text": "测试题目：缺少字段自动补全" },
        "structure": { "options": [{"id":"A", "text":"选项A"}] }
    }]
    with open(INPUT_FILE, 'w', encoding='utf-8') as f:
        json.dump(dummy_data, f, ensure_ascii=False, indent=2)
    print(f"✅ 测试文件 {INPUT_FILE} 已生成，请再次运行脚本查看效果。")

if __name__ == "__main__":
    main()