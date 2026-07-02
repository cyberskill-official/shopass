from datetime import date, timedelta


def is_double_date(d: date) -> bool:
    """True đúng khi ngày trùng tháng (1.1, 2.2, ... 12.12)."""
    return d.day == d.month


def days_to_next_double_date(d: date) -> int:
    """Số ngày tới ngày double-date kế tiếp; 0 nếu chính ngày đó là double-date."""
    probe = d
    for _ in range(366):
        if is_double_date(probe):
            return (probe - d).days
        probe += timedelta(days=1)
    return 0  # không xảy ra (luôn có double-date trong 1 năm)


def is_payday_window(d: date) -> bool:
    """True trong cửa sổ lương: quanh ngày 1 (cuối tháng trước -> ngày 2) và ngày 15 (14 -> 16)."""
    if d.day in (1, 2) or d.day in (14, 15, 16):
        return True
    nxt = d + timedelta(days=1)
    return nxt.day == 1  # ngày cuối tháng (liền trước mùng 1)
