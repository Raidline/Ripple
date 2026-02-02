package com.bmw.trip.dips.infrastructure.repository;

import com.bmw.trip.dips.infrastructure.entity.InspectionZone;
import io.quarkus.mongodb.panache.PanacheMongoRepositoryBase;
import jakarta.enterprise.context.ApplicationScoped;
import java.util.List;

@ApplicationScoped
public class InspectionZoneRepository
    implements PanacheMongoRepositoryBase<InspectionZone, String>
{

    public List<InspectionZone> findAllByIds(List<String> ids) {
        return find("_id in ?1", ids).list();
    }
}
