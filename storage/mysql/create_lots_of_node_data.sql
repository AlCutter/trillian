insert into Trees(TreeId) VALUES (0);

delimiter $$

drop procedure if exists fill_nodes $$
create procedure fill_nodes()
deterministic
begin
  declare counter int default 1;
  insert into Node values (0, "N", '', 0);
  while counter < 1000000
  do
      insert into Node (TreeId, NodeId, NodeRevision)
          select 0, concat(NodeId, counter,"-"), second(now()) from Node;
      select counter*2 into counter;
      select counter;
  end while;
end $$
delimiter ;

call fill_nodes();

