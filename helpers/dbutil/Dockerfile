FROM postgres
RUN apt-get update \
        && packages='build-essential libicu-dev postgresql-server-dev-all git' \
        && full_packages="$(apt-get install --dry-run $packages | grep ^Inst | awk '{print $2}')" \
        && apt-get install -y $packages \
        && git clone https://github.com/okbob/plpgsql_check && cd plpgsql_check \
        && make clean && make install \
        && cd .. && rm -rf plpgsql_check \
        && git clone https://github.com/theory/pgtap && cd pgtap \
        && cpan TAP::Parser::SourceHandler::pgTAP \
        && make clean && make install \
        && cd .. && rm -rf pgtap \
        && rm -rf /var/lib/apt/lists/* \
        && apt-get purge --auto-remove -y $full_packages
CMD ["postgres"]
