from datetime import date
from bottom.regressors import is_double_date, days_to_next_double_date, is_payday_window


def test_is_double_date():
    assert is_double_date(date(2026, 2, 2))    # 2.2
    assert is_double_date(date(2026, 12, 12))  # 12.12
    assert not is_double_date(date(2026, 4, 3))  # 3.4 không trùng


def test_days_to_next_double_date():
    assert days_to_next_double_date(date(2026, 12, 12)) == 0      # chính ngày
    assert days_to_next_double_date(date(2026, 12, 1)) == 11      # tới 12.12
    assert days_to_next_double_date(date(2026, 12, 13)) > 0       # vòng sang 1.1 năm sau


def test_is_payday_window():
    assert is_payday_window(date(2026, 6, 1))    # đầu tháng
    assert is_payday_window(date(2026, 6, 15))   # giữa tháng
    assert is_payday_window(date(2026, 6, 30))   # cuối tháng (liền mùng 1)
    assert not is_payday_window(date(2026, 6, 8))  # ngoài cửa sổ
