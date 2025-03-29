-- триггер для увеличения счетчика лайков

CREATE OR REPLACE FUNCTION increment_like_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE flow
    SET like_count = like_count + 1
    WHERE id = NEW.flow_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_increment_like_count
AFTER INSERT ON flow_like
FOR EACH ROW
EXECUTE FUNCTION increment_like_count();

-- триггер для уменьшения счетчика лайков

CREATE OR REPLACE FUNCTION decrement_like_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE flow
    SET like_count = like_count - 1
    WHERE id = OLD.flow_id;

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_decrement_like_count
AFTER DELETE ON flow_like
FOR EACH ROW
EXECUTE FUNCTION decrement_like_count();
