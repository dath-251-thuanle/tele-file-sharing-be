CREATE FUNCTION authorize_password(p_hash_link VARCHAR, p_password VARCHAR)
RETURNS BOOLEAN AS $$
DECLARE
    v_password_hash VARCHAR;
BEGIN
    SELECT shares(password_hash) INTO v_password_hash
    FROM shares
    WHERE hash_link = p_hash_link;
    --Change the given password to its hash and compare with the stored hash
    RETURN v_password_hash = crypt(p_password, v_password_hash); --Need to aks about the exact crypt function