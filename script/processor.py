import json
import base64
import io
import matplotlib.pyplot as plt
import numpy as np
import traceback

# ================= 配置区域 =================
INPUT_FILE = 'raw_questions.json'
OUTPUT_FILE = 'questions.json'

# 是否在成功生成图片后清空代码字段 (减小JSON体积)
CLEAR_CODE_ON_SUCCESS = True

# ================= 核心工具：绘图与节点处理 =================

def execute_code_to_image(code_str):
    """
    执行绘图代码，返回 (Base64字符串, 是否出错, 错误信息)
    """
    if not code_str or not isinstance(code_str, str) or 'plt.' not in code_str:
        return None, False, "No plotting code provided"

    try:
        # 清理画布，防止上一张图残留
        plt.clf()
        plt.close('all')

        # 设置默认配置，避免部分环境中 LaTeX 缺失导致的报错
        plt.rcParams.update({'text.usetex': False})

        # 准备执行环境
        exec_globals = {'plt': plt, 'np': np}
        exec(code_str, exec_globals)

        # 保存图片到内存
        buf = io.BytesIO()
        plt.savefig(buf, format='png', bbox_inches='tight', dpi=100)
        buf.seek(0)

        # 转 Base64
        img_b64 = "data:image/png;base64," + base64.b64encode(buf.read()).decode('utf-8')
        buf.close()

        return img_b64, False, None

    except Exception:
        error_msg = traceback.format_exc()
        return None, True, error_msg

def process_rich_node(node, context_info=""):
    """
    [原子操作] 处理单个 RichContent 节点
    遍历检查是否有 code 需要执行并生成 image
    """
    if not isinstance(node, dict):
        return

    # 逻辑判断：只有当标识为需要图片 (has_image) 且 目前没有图片数据 (image 为 null) 时才执行
    if node.get('has_image') is True and not node.get('image'):
        # 只有存在 code 时才尝试生成
        if node.get('code'):
            print(f"    🎨 [绘图] {context_info} ...")
            image_data, is_error, err_msg = execute_code_to_image(node.get('code'))

            if is_error:
                node['code_error'] = True
                node['debug_msg'] = err_msg
                print(f"    ❌ {context_info} 绘图失败")
            else:
                node['image'] = image_data
                node['code_error'] = False
                node['debug_msg'] = None
                if CLEAR_CODE_ON_SUCCESS:
                    node['code'] = None # 清空代码以节省空间
                print(f"    ✅ {context_info} 生成成功")
        else:
            # 有意图但无代码的情况标记为错误
            node['code_error'] = True
            node['debug_msg'] = "has_image is true but code is missing."
    else:
        # 确保基础字段存在，方便前端处理
        if 'code_error' not in node:
            node['code_error'] = False

# ================= 题型特定策略 (Handlers) =================

def process_common_parts(question):
    """
    处理所有题型通用的部分：题干(content) 和 解析(validation.explanation)
    """
    q_id = question.get('id', 'Unknown')

    # 1. 处理题干
    if 'content' in question:
        process_rich_node(question['content'], f"题目[{q_id}]-题干")

    # 2. 处理解析
    if 'validation' in question and 'explanation' in question['validation']:
        process_rich_node(question['validation']['explanation'], f"题目[{q_id}]-解析")

def handle_choice_style_question(question):
    """
    处理 [选择类] 题目 (单选、多选、判断)
    特点：structure 中包含 options 数组
    """
    process_common_parts(question)

    # 处理选项中的图片 (虽然判断题选项通常只有文字，但保留此逻辑兼容性更好)
    options = question.get('structure', {}).get('options', [])
    q_id = question.get('id')
    for opt in options:
        process_rich_node(opt, f"题目[{q_id}]-选项[{opt.get('id')}]")

def handle_cloze_question(question):
    """
    处理 [完形填空] 题目
    特点：structure 中包含 blanks 数组，每个 blank 里有 options
    """
    process_common_parts(question)

    blanks = question.get('structure', {}).get('blanks', [])
    q_id = question.get('id')

    for blank in blanks:
        blank_id = blank.get('id')
        options = blank.get('options', [])
        for opt in options:
            process_rich_node(opt, f"题目[{q_id}]-空({blank_id})-选项[{opt.get('id')}]")

def handle_basic_question(question):
    """
    处理 [基础类] 题目 (简答、普通填空)
    特点：没有复杂的选项结构，只需处理通用部分
    """
    process_common_parts(question)

# ================= 路由分发 (Router) =================

# 将题型映射到对应的处理函数
PROCESSOR_MAP = {
    'single_choice': handle_choice_style_question,   # 单选
    'multiple_choice': handle_choice_style_question, # 多选
    'true_false': handle_choice_style_question,      # <--- 新增：判断题 (结构类似选择题)
    'short_answer': handle_basic_question,           # 简答
    'fill_blank': handle_basic_question,             # 填空
    'cloze': handle_cloze_question                   # 完形填空
}

def dispatch_processor(question):
    q_type = question.get('type')
    handler = PROCESSOR_MAP.get(q_type)

    if handler:
        handler(question)
    else:
        print(f"⚠️ 未知的题目类型: {q_type}, 仅处理通用部分(题干/解析)")
        process_common_parts(question)

# ================= 主程序 =================

def main():
    print(f"📂 正在读取数据源: {INPUT_FILE} ...")
    try:
        with open(INPUT_FILE, 'r', encoding='utf-8') as f:
            questions = json.load(f)
    except Exception as e:
        print(f"❌ 读取文件失败: {e}")
        return

    if not isinstance(questions, list):
        print("❌ JSON 格式错误：根节点应当是一个数组")
        return

    total = len(questions)
    print(f"⚙️ 开始处理 {total} 道题目...")

    for i, q in enumerate(questions):
        print(f"[{i+1}/{total}] 处理 ID: {q.get('id')} | 类型: {q.get('type')}")
        dispatch_processor(q)

    print(f"💾 正在保存结果到: {OUTPUT_FILE} ...")
    try:
        with open(OUTPUT_FILE, 'w', encoding='utf-8') as f:
            json.dump(questions, f, ensure_ascii=False, indent=2)
        print("✨ 处理程序运行结束！")
    except Exception as e:
        print(f"❌ 保存文件失败: {e}")

if __name__ == '__main__':
    main()