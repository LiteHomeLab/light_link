"""LightLink 参数验证器"""
from typing import Dict, Any, List
from dataclasses import dataclass, field
from lightlink.metadata import MethodMetadata


@dataclass
class ValidationError:
    """验证错误"""
    parameter_name: str
    expected_type: str
    actual_type: str
    actual_value: Any = None
    message: str = ""


@dataclass
class ValidationResult:
    """验证结果"""
    is_valid: bool = True
    error_message: str = ""
    errors: List[ValidationError] = field(default_factory=list)


class Validator:
    """参数验证器"""

    def __init__(self, method_meta: MethodMetadata):
        self.method_meta = method_meta

    def validate(self, args: Dict[str, Any]) -> ValidationResult:
        """验证参数"""
        result = ValidationResult(is_valid=True)

        for param_meta in self.method_meta.params:
            if param_meta.required and param_meta.name not in args:
                result.is_valid = False
                result.errors.append(ValidationError(
                    parameter_name=param_meta.name,
                    expected_type=param_meta.type,
                    actual_type="missing",
                    message=f"Required parameter '{param_meta.name}' is missing"
                ))
                continue

            if param_meta.name not in args:
                continue

            value = args[param_meta.name]
            actual_type = self._infer_type(value)

            if not self._is_type_compatible(param_meta.type, actual_type):
                result.is_valid = False
                result.errors.append(ValidationError(
                    parameter_name=param_meta.name,
                    expected_type=param_meta.type,
                    actual_type=actual_type,
                    actual_value=value,
                    message=f"Parameter '{param_meta.name}': expected type {param_meta.type}, got {actual_type}"
                ))

        if result.errors:
            result.error_message = f"Validation failed with {len(result.errors)} error(s)"

        return result

    @staticmethod
    def _infer_type(value: Any) -> str:
        """推断值的类型"""
        if value is None:
            return "null"
        if isinstance(value, bool):
            return "boolean"
        if isinstance(value, (int, float)):
            return "number"
        if isinstance(value, str):
            return "string"
        if isinstance(value, list):
            return "array"
        if isinstance(value, dict):
            return "object"
        return "unknown"

    @staticmethod
    def _is_type_compatible(expected: str, actual: str) -> bool:
        """检查类型是否兼容"""
        return expected == actual
