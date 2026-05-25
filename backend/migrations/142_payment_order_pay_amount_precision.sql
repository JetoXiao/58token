ALTER TABLE payment_orders
    ALTER COLUMN pay_amount TYPE DECIMAL(20,8)
    USING pay_amount::DECIMAL(20,8);
