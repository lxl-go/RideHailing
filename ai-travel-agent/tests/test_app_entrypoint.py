import ast
from pathlib import Path
import unittest


class AppEntrypointTest(unittest.TestCase):
    def test_app_supports_python_file_startup(self):
        source = Path(__file__).resolve().parents[1].joinpath("app.py").read_text(encoding="utf-8")
        tree = ast.parse(source)

        main_blocks = [
            node
            for node in tree.body
            if isinstance(node, ast.If)
            and isinstance(node.test, ast.Compare)
            and ast.unparse(node.test) == "__name__ == '__main__'"
        ]

        self.assertEqual(1, len(main_blocks))
        calls = [node for node in ast.walk(main_blocks[0]) if isinstance(node, ast.Call)]
        self.assertTrue(any(ast.unparse(call.func) == "uvicorn.run" for call in calls))


if __name__ == "__main__":
    unittest.main()
