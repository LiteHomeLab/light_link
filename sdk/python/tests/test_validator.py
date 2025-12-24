import pytest
from lightlink.validator import Validator, ValidationResult
from lightlink.metadata import MethodMetadata, ParameterMetadata


def test_validate_valid_params():
    meta = MethodMetadata(
        name="test",
        description="Test",
        params=[ParameterMetadata(name="a", type="number", required=True, description="")],
        returns=[]
    )
    validator = Validator(meta)
    result = validator.validate({"a": 42})

    assert result.is_valid


def test_validate_missing_required():
    meta = MethodMetadata(
        name="test",
        description="Test",
        params=[ParameterMetadata(name="a", type="number", required=True, description="")],
        returns=[]
    )
    validator = Validator(meta)
    result = validator.validate({})

    assert not result.is_valid
    assert "missing" in result.errors[0].actual_type
