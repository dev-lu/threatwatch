#!/bin/bash
echo "########################Starting to execute SH script...########################"

while ! cqlsh logging_db -u "${CASSANDRA_USER}" -p "${CASSANDRA_USER}" -e 'describe cluster' ; do
     echo "########################Waiting for main instance to be ready...########################"
     sleep 5
done

for cql_file in ./tmp/cql/*.cql;
do
  cqlsh logging_db -u "${CASSANDRA_USER}" -p "${CASSANDRA_USER}" -f "${cql_file}" ;
  echo "########################Script ""${cql_file}"" executed!!!########################"
done
echo "########################Execution of SH script is finished!########################"
echo "########################Stopping temporary instance!########################"