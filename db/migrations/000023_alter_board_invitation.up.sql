ALTER TABLE board_invitation DROP CONSTRAINT board_invitation_board_id_fkey;
ALTER TABLE board_invitation ADD CONSTRAINT board_invitation_board_id_fkey FOREIGN KEY (board_id) REFERENCES board(id) ON DELETE CASCADE;
